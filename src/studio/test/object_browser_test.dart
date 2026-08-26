import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

import 'package:qtcloud_studio/main.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  test('Bucket category, browsing, accent, and object size labels are stable',
      () {
    const studio = Bucket(
      name: 'qtcloud-asset-studio',
      region: 'oss-cn-hangzhou',
      storageClass: 'Standard',
      createdAt: '2026-08-25',
    );
    const private = Bucket(
      name: 'qtadmin-private',
      region: 'oss-cn-hangzhou',
      storageClass: 'Standard',
      createdAt: '2026-08-24',
    );
    const site = Bucket(
      name: 'qtdocs-site',
      region: 'oss-cn-hangzhou',
      storageClass: 'Standard',
      createdAt: '2026-08-23',
    );
    const provider = Bucket(
      name: 'qtadmin-provider',
      region: 'oss-cn-hangzhou',
      storageClass: 'Standard',
      createdAt: '2026-08-22',
    );
    const other = Bucket(
      name: 'qtcloud-data',
      region: 'oss-cn-hangzhou',
      storageClass: 'Standard',
      createdAt: '2026-08-21',
    );

    expect(studio.category, 'Studio');
    expect(private.category, 'Private');
    expect(site.category, 'Site');
    expect(provider.category, 'Provider');
    expect(other.category, 'Other');
    expect(studio.canBrowseObjects, isTrue);
    expect(private.canBrowseObjects, isFalse);
    expect(studio.categoryIcon, Icons.web);
    expect(private.categoryIcon, Icons.lock);
    expect(site.categoryIcon, Icons.public);
    expect(provider.categoryIcon, Icons.dns);
    expect(other.categoryIcon, Icons.storage);
    expect(studio.categoryColor, isA<Color>());
    expect(other.accentColor, isA<Color>());

    expect(
      visibleBucketsForRole('viewer', [studio, private]).map((b) => b.name),
      ['qtcloud-asset-studio'],
    );
    expect(
      visibleBucketsForRole('admin', [studio, private]).map((b) => b.name),
      ['qtcloud-asset-studio', 'qtadmin-private'],
    );

    expect(
      OssObject.fromJson(const {
        'key': 'folder/',
        'size': 0,
        'type': 'Directory',
      }).isDir,
      isTrue,
    );
    expect(
      const OssObject(
        key: 'small.txt',
        size: 512,
        type: 'Normal',
        storageClass: 'Standard',
        lastModified: '',
      ).sizeLabel,
      '512 B',
    );
    expect(
      const OssObject(
        key: 'kb.txt',
        size: 2048,
        type: 'Normal',
        storageClass: 'Standard',
        lastModified: '',
      ).sizeLabel,
      '2.00 KB',
    );
    expect(
      const OssObject(
        key: 'mb.bin',
        size: 3 * 1024 * 1024,
        type: 'Normal',
        storageClass: 'Standard',
        lastModified: '',
      ).sizeLabel,
      '3.00 MB',
    );
    expect(
      const OssObject(
        key: 'gb.bin',
        size: 2 * 1024 * 1024 * 1024,
        type: 'Normal',
        storageClass: 'Standard',
        lastModified: '',
      ).sizeLabel,
      '2.00 GB',
    );
  });

  test('object pagination stops when provider repeats the same marker',
      () async {
    final requestedMarkers = <String>[];

    final objects = await fetchAllObjectPages((marker) async {
      requestedMarkers.add(marker);
      return {
        'objects': [
          {
            'key': marker.isEmpty ? 'first.txt' : 'repeat.txt',
            'size': 1,
          },
        ],
        'truncated': true,
        'next_marker': marker.isEmpty ? 'same' : 'same',
      };
    });

    expect(requestedMarkers, ['', 'same']);
    expect(objects.map((object) => object.key), ['first.txt', 'repeat.txt']);
  });

  testWidgets(
      'BucketObjectsScreen browses folders, searches, sorts, and copies',
      (WidgetTester tester) async {
    final requested = <String>[];
    final clipboardWrites = <String>[];
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

    final client = _objectClient((request) async {
      requested
          .add('${request.method} ${request.url.path}?${request.url.query}');
      if (request.url.path.endsWith('/object-url')) {
        expect(request.url.queryParameters['key'], 'folder/nested.txt');
        expect(request.url.queryParameters['expires'], '86400');
        return http.Response(
            '{"url":"https://oss.example.com/nested.txt"}', 200);
      }
      return http.Response(_objectsBody, 200);
    });

    await tester.pumpWidget(
      MaterialApp(
        home: BucketObjectsScreen(
          bucketName: 'qtcloud-asset-studio',
          client: client,
        ),
      ),
    );
    await _pumpAsync(tester);

    expect(find.text('root.txt'), findsOneWidget);
    expect(find.text('folder/'), findsOneWidget);
    expect(find.text('folder/nested.txt'), findsNothing);

    await tester.tap(find.text('folder/'));
    await _pumpAsync(tester);

    expect(find.text('返回上级'), findsOneWidget);
    expect(find.text('folder/nested.txt'), findsOneWidget);
    expect(find.text('folder/z-last.txt'), findsOneWidget);

    await tester.enterText(find.byType(TextField), 'nested');
    await _pumpAsync(tester);
    expect(find.text('folder/nested.txt'), findsOneWidget);
    expect(find.text('folder/z-last.txt'), findsNothing);

    await tester.enterText(find.byType(TextField), 'missing');
    await _pumpAsync(tester);
    expect(find.text('没有匹配的文件'), findsOneWidget);

    await tester.enterText(find.byType(TextField), '');
    await _pumpAsync(tester);

    await tester.tap(find.text('大小'));
    await _pumpAsync(tester);
    expect(
      tester.getTopLeft(find.text('folder/z-last.txt')).dy,
      lessThan(tester.getTopLeft(find.text('folder/nested.txt')).dy),
    );

    await tester.tap(find.text('大小'));
    await _pumpAsync(tester);
    expect(
      tester.getTopLeft(find.text('folder/nested.txt')).dy,
      lessThan(tester.getTopLeft(find.text('folder/z-last.txt')).dy),
    );

    await tester.tap(find.text('日期'));
    await _pumpAsync(tester);
    expect(
      tester.getTopLeft(find.text('folder/z-last.txt')).dy,
      lessThan(tester.getTopLeft(find.text('folder/nested.txt')).dy),
    );

    await tester.tap(find.text('日期'));
    await _pumpAsync(tester);
    expect(
      tester.getTopLeft(find.text('folder/nested.txt')).dy,
      lessThan(tester.getTopLeft(find.text('folder/z-last.txt')).dy),
    );

    await tester.tap(find.byTooltip('复制链接').first);
    await _pumpAsync(tester);
    expect(find.text('选择链接有效期'), findsOneWidget);

    await tester.tap(find.text('1 天'));
    await _pumpAsync(tester);

    expect(clipboardWrites, ['https://oss.example.com/nested.txt']);

    await tester.tap(find.text('返回上级'));
    await _pumpAsync(tester);
    expect(find.text('root.txt'), findsOneWidget);
    expect(requested.any((item) => item.contains('/object-url')), isTrue);
  });

  testWidgets('BucketObjectsScreen keeps metadata-only buckets linkless',
      (WidgetTester tester) async {
    await tester.pumpWidget(
      MaterialApp(
        home: BucketObjectsScreen(
          bucketName: 'qtadmin-private',
          client: _objectClient((_) async => http.Response(_objectsBody, 200)),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('root.txt'), findsOneWidget);
    expect(find.byTooltip('复制链接'), findsNothing);
  });

  testWidgets('BucketObjectsScreen renders provider errors',
      (WidgetTester tester) async {
    await tester.pumpWidget(
      MaterialApp(
        home: BucketObjectsScreen(
          bucketName: 'qtcloud-asset-studio',
          client: _objectClient(
            (_) async => http.Response('{"error":"denied"}', 403),
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.textContaining('加载文件失败'), findsOneWidget);
    expect(find.textContaining('denied'), findsOneWidget);
  });

  testWidgets('BucketObjectsScreen reports copy link failures',
      (WidgetTester tester) async {
    final client = _objectClient((request) async {
      if (request.url.path.endsWith('/object-url')) {
        return http.Response('{"error":"private bucket"}', 403);
      }
      return http.Response(_objectsBody, 200);
    });

    await tester.pumpWidget(
      MaterialApp(
        home: BucketObjectsScreen(
          bucketName: 'qtcloud-asset-studio',
          client: client,
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('folder/'));
    await tester.pumpAndSettle();
    await tester.tap(find.byTooltip('复制链接').first);
    await tester.pumpAndSettle();
    await tester.tap(find.text('7 天'));
    await tester.pumpAndSettle();

    expect(find.textContaining('生成链接失败'), findsOneWidget);
    expect(find.textContaining('private bucket'), findsOneWidget);
  });
}

ProviderApiClient _objectClient(
  Future<http.Response> Function(http.Request request) handler,
) {
  return ProviderApiClient(
    baseUrl: 'https://api.example.com',
    httpClient: MockClient(handler),
  );
}

Future<void> _pumpAsync(WidgetTester tester) async {
  await tester.pump();
  await tester.pump(const Duration(milliseconds: 50));
}

const _objectsBody = '''
{
  "objects": [
    {
      "key": "root.txt",
      "size": 512,
      "type": "Normal",
      "storage_class": "Standard",
      "last_modified": "2026-08-24 09:00:00"
    },
    {
      "key": "folder/nested.txt",
      "size": 1024,
      "type": "Normal",
      "storage_class": "Standard",
      "last_modified": "2026-08-24 10:00:00"
    },
    {
      "key": "folder/z-last.txt",
      "size": 2048,
      "type": "Normal",
      "storage_class": "Standard",
      "last_modified": "2026-08-24 11:00:00"
    }
  ],
  "truncated": false,
  "next_marker": ""
}
''';
