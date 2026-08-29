import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

import 'package:qtcloud_studio/main.dart';

void main() {
  test('shareTokenFromUri recognizes path and hash share routes', () {
    expect(
      shareTokenFromUri(Uri.parse('https://asset.example.com/share/abc123')),
      'abc123',
    );
    expect(
      shareTokenFromUri(Uri.parse('https://asset.example.com/#/share/hash123')),
      'hash123',
    );
    expect(shareTokenFromUri(Uri.parse('https://asset.example.com/')), isNull);
    expect(
      shareTokenFromUri(Uri.parse('https://asset.example.com/share/')),
      isNull,
    );
  });

  testWidgets('PublicShareScreen loads without authentication', (tester) async {
    final requested = <String>[];
    final client = ProviderApiClient(
      baseUrl: 'https://api.example.com',
      httpClient: MockClient((request) async {
        requested.add(request.url.path);
        if (request.url.path == '/shares/share-token') {
          return http.Response.bytes(
            utf8.encode(_shareBody),
            200,
            headers: {'content-type': 'application/json; charset=utf-8'},
          );
        }
        return http.Response.bytes(
          utf8.encode(_sharedObjectsBody),
          200,
          headers: {'content-type': 'application/json; charset=utf-8'},
        );
      }),
    );

    await tester.pumpWidget(
      MaterialApp(
        home: PublicShareScreen(token: 'share-token', client: client),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('公开资料'), findsOneWidget);
    expect(find.text('readme.md'), findsOneWidget);
    expect(find.textContaining('只读分享'), findsOneWidget);
    expect(requested, [
      '/shares/share-token',
      '/shares/share-token/objects',
    ]);
    expect(find.text('登录'), findsNothing);
  });

  testWidgets('PublicShareScreen downloads a shared file', (tester) async {
    final requested = <String>[];
    final downloads = <String>[];
    final client = ProviderApiClient(
      baseUrl: 'https://api.example.com',
      httpClient: MockClient((request) async {
        requested.add(request.url.toString());
        if (request.url.path == '/shares/share-token') {
          return http.Response.bytes(
            utf8.encode(_shareBody),
            200,
            headers: {'content-type': 'application/json; charset=utf-8'},
          );
        }
        if (request.url.path.endsWith('/objects')) {
          return http.Response.bytes(
            utf8.encode(_sharedObjectsBody),
            200,
            headers: {'content-type': 'application/json; charset=utf-8'},
          );
        }
        return http.Response.bytes(
          utf8.encode(
            '{"url":"https://oss.example.com/docs/readme.md","expires_in":0}',
          ),
          200,
          headers: {'content-type': 'application/json; charset=utf-8'},
        );
      }),
    );

    await tester.pumpWidget(
      MaterialApp(
        home: PublicShareScreen(
          token: 'share-token',
          client: client,
          onDownload: (url, fileName) async {
            downloads.add('${url.toString()}::$fileName');
          },
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.byTooltip('下载文件'));
    await tester.pumpAndSettle();

    expect(
      requested.last,
      'https://api.example.com/shares/share-token/object-url?key=docs%2Freadme.md',
    );
    expect(downloads, [
      'https://oss.example.com/docs/readme.md::readme.md',
    ]);
    expect(find.text('下载已开始'), findsOneWidget);
  });

  testWidgets('PublicShareScreen downloads all shared files', (tester) async {
    final downloads = <String>[];
    final objectRequests = <String>[];
    final client = ProviderApiClient(
      baseUrl: 'https://api.example.com',
      httpClient: MockClient((request) async {
        if (request.url.path == '/shares/share-token') {
          return http.Response.bytes(
            utf8.encode(_shareBody),
            200,
            headers: {'content-type': 'application/json; charset=utf-8'},
          );
        }
        if (request.url.path.endsWith('/objects')) {
          return http.Response.bytes(
              utf8.encode(_multipleSharedObjectsBody), 200);
        }
        if (request.url.path.endsWith('/object-url')) {
          objectRequests.add(request.url.queryParameters['key']!);
          final key = request.url.queryParameters['key']!;
          return http.Response.bytes(
            utf8.encode(
              '{"url":"https://oss.example.com/$key","expires_in":0}',
            ),
            200,
          );
        }
        if (request.url.host == 'oss.example.com') {
          final key = request.url.path.substring(1);
          return http.Response.bytes(utf8.encode('content:$key'), 200);
        }
        return http.Response('unexpected request', 500);
      }),
    );

    await tester.pumpWidget(
      MaterialApp(
        home: PublicShareScreen(
          token: 'share-token',
          client: client,
          onDownloadBytes: (bytes, fileName) async {
            downloads.add('${bytes.length}::$fileName');
          },
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.byTooltip('下载全部'));
    await tester.pumpAndSettle();

    expect(objectRequests, ['docs/readme.md', 'docs/guide.pdf']);
    expect(downloads.single.endsWith('::公开资料.zip'), isTrue);
    expect(int.parse(downloads.single.split('::').first), greaterThan(4));
    expect(find.text('已开始下载：公开资料.zip'), findsOneWidget);
  });

  testWidgets('PublicShareScreen renders revoked share state', (tester) async {
    final client = ProviderApiClient(
      baseUrl: 'https://api.example.com',
      httpClient: MockClient(
        (_) async => http.Response.bytes(
          utf8.encode('{"error":"share not found"}'),
          404,
          headers: {'content-type': 'application/json; charset=utf-8'},
        ),
      ),
    );

    await tester.pumpWidget(
      MaterialApp(
        home: PublicShareScreen(token: 'revoked', client: client),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('分享链接不存在或已被撤销'), findsOneWidget);
  });

  testWidgets('MySharesScreen lists shares and supports copy and revoke',
      (tester) async {
    final requests = <String>[];
    final clipboardWrites = <String>[];
    var revoked = false;

    TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger
        .setMockMethodCallHandler(SystemChannels.platform, (call) async {
      if (call.method == 'Clipboard.setData') {
        final arguments = Map<Object?, Object?>.from(call.arguments as Map);
        clipboardWrites.add(arguments['text'] as String);
      }
      return null;
    });
    addTearDown(() {
      TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger
          .setMockMethodCallHandler(SystemChannels.platform, null);
    });

    final client = ProviderApiClient(
      baseUrl: 'https://api.example.com',
      httpClient: MockClient((request) async {
        requests.add('${request.method} ${request.url.path}');
        if (request.method == 'DELETE' &&
            request.url.path == '/shares/share-token') {
          revoked = true;
          return http.Response('', 204);
        }
        if (request.method == 'GET' && request.url.path == '/shares') {
          return http.Response.bytes(
            utf8.encode(revoked ? _emptySharesBody : _sharesBody),
            200,
            headers: {'content-type': 'application/json; charset=utf-8'},
          );
        }
        return http.Response('{"error":"unexpected"}', 500);
      }),
    );

    await tester.pumpWidget(
      MaterialApp(
        home: MySharesScreen(client: client),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('我的分享'), findsOneWidget);
    expect(find.text('公开资料'), findsOneWidget);
    expect(find.text('docs/'), findsOneWidget);

    await tester.tap(find.byTooltip('复制分享链接'));
    await tester.pumpAndSettle();
    expect(clipboardWrites, [
      'https://asset.cloud.quanttide.com/#/share/share-token',
    ]);
    expect(find.text('分享链接已复制'), findsOneWidget);

    await tester.tap(find.byTooltip('撤销分享'));
    await tester.pumpAndSettle();
    expect(find.text('确认撤销分享'), findsOneWidget);
    expect(requests, ['GET /shares']);

    await tester.tap(find.text('撤销'));
    await tester.pumpAndSettle();

    expect(requests, [
      'GET /shares',
      'DELETE /shares/share-token',
      'GET /shares',
    ]);
    expect(find.text('暂无分享链接'), findsOneWidget);
    expect(find.text('分享已撤销'), findsOneWidget);
  });
}

const _shareBody = '''
{
  "share": {
    "token": "share-token",
    "title": "公开资料",
    "bucket": "qtcloud-asset-studio",
    "prefixes": ["docs/"],
    "url": "https://asset.cloud.quanttide.com/#/share/share-token",
    "created_at": "2026-08-27T10:00:00Z"
  }
}
''';

const _sharedObjectsBody = '''
{
  "bucket": "qtcloud-asset-studio",
  "objects": [
    {
      "key": "docs/readme.md",
      "size": 128,
      "type": "Normal",
      "storage_class": "Standard",
      "last_modified": "2026-08-27 10:00:00"
    }
  ],
  "truncated": false,
  "next_marker": ""
}
''';

const _multipleSharedObjectsBody = '''
{
  "bucket": "qtcloud-asset-studio",
  "objects": [
    {
      "key": "docs/readme.md",
      "size": 128,
      "type": "Normal",
      "storage_class": "Standard",
      "last_modified": "2026-08-27 10:00:00"
    },
    {
      "key": "docs/guide.pdf",
      "size": 256,
      "type": "Normal",
      "storage_class": "Standard",
      "last_modified": "2026-08-27 10:01:00"
    }
  ],
  "truncated": false,
  "next_marker": ""
}
''';

const _sharesBody = '''
{
  "shares": [
    {
      "token": "share-token",
      "title": "公开资料",
      "bucket": "qtcloud-asset-studio",
      "prefixes": ["docs/"],
      "url": "https://asset.cloud.quanttide.com/#/share/share-token",
      "created_at": "2026-08-27T10:00:00Z"
    }
  ],
  "total": 1
}
''';

const _emptySharesBody = '''
{
  "shares": [],
  "total": 0
}
''';
