import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

import 'package:qtcloud_studio/main.dart';

void main() {
  test('share archive timeout allows the production archive window', () {
    expect(shareArchiveTimeout, const Duration(seconds: 300));
  });

  test('ProviderApiClient trims base URL and exposes login URI', () {
    final client = ProviderApiClient(
      baseUrl: 'https://api.example.com///',
      httpClient: MockClient((_) async => http.Response('{}', 200)),
    );

    expect(client.baseUrl, 'https://api.example.com');
    expect(client.loginUri, Uri.parse('https://api.example.com/auth/login'));
  });

  test('ProviderApiClient fetches current user and falls back to email display',
      () async {
    final client = ProviderApiClient(
      baseUrl: 'https://api.example.com',
      httpClient: MockClient((request) async {
        expect(request.url.path, '/auth/me');
        return http.Response(_userWithoutNameBody, 200);
      }),
    );

    final user = await client.fetchCurrentUser();

    expect(user.email, 'viewer@example.com');
    expect(user.displayName, 'viewer@example.com');
  });

  test('ProviderApiClient logs in with local account credentials', () async {
    final client = ProviderApiClient(
      baseUrl: 'https://api.example.com',
      httpClient: MockClient((request) async {
        expect(request.method, 'POST');
        expect(request.url.path, '/auth/login');
        expect(request.headers['Content-Type'], contains('application/json'));
        expect(jsonDecode(request.body), {
          'account': 'admin',
          'password': 'correct-password',
        });
        return http.Response(_adminUserBody, 200);
      }),
    );

    final user = await client.login(
      account: 'admin',
      password: 'correct-password',
    );

    expect(user.account, 'admin');
    expect(user.role, 'admin');
  });

  test('ProviderApiClient maps 401 current user responses to auth required',
      () async {
    final client = ProviderApiClient(
      baseUrl: 'https://api.example.com',
      httpClient: MockClient(
        (_) async => http.Response('{"error":"auth required"}', 401),
      ),
    );

    expect(client.fetchCurrentUser(), throwsA(isA<AuthRequiredException>()));
    expect(const AuthRequiredException().toString(), 'authentication required');
  });

  test('ProviderApiClient treats logout 401 as already logged out', () async {
    final client = ProviderApiClient(
      baseUrl: 'https://api.example.com',
      httpClient: MockClient((request) async {
        expect(request.method, 'POST');
        expect(request.url.path, '/auth/logout');
        return http.Response('{"error":"auth required"}', 401);
      }),
    );

    expect(client.logout(), completes);
  });

  test('ProviderApiClient parses buckets and health responses', () async {
    final requested = <String>[];
    final client = ProviderApiClient(
      baseUrl: 'https://api.example.com',
      httpClient: MockClient((request) async {
        requested.add(request.url.path);
        return switch (request.url.path) {
          '/health' => http.Response(_healthBody, 200),
          '/buckets' => http.Response(_bucketsBody, 200),
          _ => http.Response('{"error":"unexpected"}', 500),
        };
      }),
    );

    expect(await client.fetchHealth(), containsPair('status', 'ok'));
    final buckets = await client.fetchBuckets();

    expect(requested, ['/health', '/buckets']);
    expect(buckets.map((bucket) => bucket.name), [
      'qtcloud-asset-studio',
      'qtadmin-private',
    ]);
  });

  test('ProviderApiClient follows object pages and encodes bucket names',
      () async {
    final requested = <String>[];
    final client = ProviderApiClient(
      baseUrl: 'https://api.example.com',
      httpClient: MockClient((request) async {
        requested.add(request.url.toString());
        expect(request.url.path, '/buckets/space%20bucket/objects');
        if (request.url.queryParameters['marker'] == 'next marker') {
          return http.Response(_objectsSecondPageBody, 200);
        }
        return http.Response(_objectsFirstPageBody, 200);
      }),
    );

    final objects = await client.fetchObjects('space bucket');

    expect(
        objects.map((object) => object.key), ['root.txt', 'folder/file.txt']);
    expect(requested.first,
        'https://api.example.com/buckets/space%20bucket/objects');
    expect(requested.last, contains('marker=next+marker'));
  });

  test('ProviderApiClient fetches object URLs with encoded query parameters',
      () async {
    final client = ProviderApiClient(
      baseUrl: 'https://api.example.com',
      httpClient: MockClient((request) async {
        expect(request.url.path, '/buckets/qtcloud-asset-studio/object-url');
        expect(request.url.queryParameters['key'], 'folder/a b.txt');
        expect(request.url.queryParameters['expires'], '86400');
        return http.Response(
            '{"url":"https://oss.example.com/folder/a%20b.txt"}', 200);
      }),
    );

    final url = await client.fetchObjectUrl(
      bucketName: 'qtcloud-asset-studio',
      objectKey: 'folder/a b.txt',
      expiresSeconds: 86400,
    );

    expect(url, 'https://oss.example.com/folder/a%20b.txt');
  });

  test('ProviderApiClient creates and reads public folder shares', () async {
    final requests = <http.Request>[];
    final client = ProviderApiClient(
      baseUrl: 'https://api.example.com',
      httpClient: MockClient((request) async {
        requests.add(request);
        if (request.method == 'POST' && request.url.path == '/shares') {
          expect(jsonDecode(request.body), {
            'title': '公开资料',
            'bucket': 'qtcloud-asset-studio',
            'prefixes': ['docs/', 'design/'],
          });
          return http.Response.bytes(utf8.encode(_shareBody), 201);
        }
        expect(request.method, 'GET');
        expect(request.url.path, '/shares/share-token');
        return http.Response.bytes(utf8.encode(_shareBody), 200);
      }),
    );

    final created = await client.createShare(
      title: '公开资料',
      bucketName: 'qtcloud-asset-studio',
      prefixes: const ['docs/', 'design/'],
    );
    final loaded = await client.fetchShare('share-token');

    expect(created.token, 'share-token');
    expect(loaded.prefixes, ['design/', 'docs/']);
    expect(requests.map((request) => '${request.method} ${request.url.path}'), [
      'POST /shares',
      'GET /shares/share-token',
    ]);
  });

  test('ProviderApiClient fetches shared objects and shared object URLs',
      () async {
    final requested = <String>[];
    final client = ProviderApiClient(
      baseUrl: 'https://api.example.com',
      httpClient: MockClient((request) async {
        requested.add(request.url.toString());
        if (request.url.path.endsWith('/objects')) {
          return http.Response(_sharedObjectsBody, 200);
        }
        return http.Response(
          '{"url":"https://oss.example.com/docs/readme.md","expires_in":0}',
          200,
        );
      }),
    );

    final objects = await client.fetchShareObjects(
      token: 'share token',
      prefix: 'docs/',
    );
    final url = await client.fetchShareObjectUrl(
      token: 'share token',
      objectKey: 'docs/readme.md',
    );

    expect(objects.map((object) => object.key), ['docs/readme.md']);
    expect(url, 'https://oss.example.com/docs/readme.md');
    expect(requested.first, contains('/shares/share%20token/objects'));
    expect(requested.first, contains('prefix=docs%2F'));
    expect(requested.last, contains('/shares/share%20token/object-url'));
  });

  test('ProviderApiClient fetches a shared ZIP archive', () async {
    final requested = <String>[];
    final client = ProviderApiClient(
      baseUrl: 'https://api.example.com',
      httpClient: MockClient((request) async {
        requested.add(request.url.toString());
        return http.Response.bytes([0x50, 0x4b, 0x03, 0x04], 200);
      }),
    );

    final archive = await client.fetchShareArchive('share token');

    expect(archive, [0x50, 0x4b, 0x03, 0x04]);
    expect(requested, [
      'https://api.example.com/shares/share%20token/download',
    ]);
  });

  test('ProviderApiClient uses a separate client for public object bytes',
      () async {
    final apiRequests = <String>[];
    final objectRequests = <String>[];
    final apiClient = MockClient((request) async {
      apiRequests.add(request.url.toString());
      return http.Response(
        '{"url":"https://oss.example.com/docs/readme.md","expires_in":0}',
        200,
      );
    });
    final objectClient = MockClient((request) async {
      objectRequests.add(request.url.toString());
      return http.Response.bytes(utf8.encode('content'), 200);
    });

    final client = ProviderApiClient(
      baseUrl: 'https://api.example.com',
      httpClient: apiClient,
      publicObjectHttpClient: objectClient,
    );
    final bytes = await client.fetchShareObjectBytes(
      token: 'share-token',
      objectKey: 'docs/readme.md',
    );

    expect(bytes, utf8.encode('content'));
    expect(apiRequests, [
      'https://api.example.com/shares/share-token/object-url?key=docs%2Freadme.md',
    ]);
    expect(objectRequests, ['https://oss.example.com/docs/readme.md']);
  });

  test('ProviderApiClient supports admin user management endpoints', () async {
    final requested = <String>[];
    final bodies = <String>[];
    final client = ProviderApiClient(
      baseUrl: 'https://api.example.com',
      httpClient: MockClient((request) async {
        requested.add('${request.method} ${request.url.path}');
        if (request.method == 'POST' || request.method == 'PATCH') {
          bodies.add(request.body);
        }
        return switch ((request.method, request.url.path)) {
          ('GET', '/admin/users') => http.Response(_usersBody, 200),
          ('POST', '/admin/users') => http.Response(_invitedUserBody, 201),
          ('PATCH', '/admin/users/user-1/role') =>
            http.Response(_updatedUserBody, 200),
          ('POST', '/admin/users/user-1/disable') => http.Response('{}', 200),
          ('POST', '/admin/users/user-1/sessions/revoke') =>
            http.Response('{}', 200),
          _ => http.Response('{"error":"unexpected"}', 500),
        };
      }),
    );

    final users = await client.fetchUsers();
    final invited = await client.inviteUser(
      account: 'new-user',
      name: 'New User',
      role: 'admin',
      password: '123456',
    );
    final updated =
        await client.updateUserRole(userId: 'user-1', role: 'admin');
    await client.disableUser('user-1');
    await client.revokeUserSessions('user-1');

    expect(users, hasLength(1));
    expect(invited.account, 'new-user');
    expect(updated.role, 'admin');
    expect(requested, [
      'GET /admin/users',
      'POST /admin/users',
      'PATCH /admin/users/user-1/role',
      'POST /admin/users/user-1/disable',
      'POST /admin/users/user-1/sessions/revoke',
    ]);
    expect(jsonDecode(bodies[0]), {
      'account': 'new-user',
      'name': 'New User',
      'role': 'admin',
      'password': '123456',
    });
    expect(bodies[1], contains('admin'));
  });

  test('ProviderApiClient surfaces provider error messages', () async {
    final client = ProviderApiClient(
      baseUrl: 'https://api.example.com',
      httpClient: MockClient(
        (_) async => http.Response(
          '{"error":"Service Unavailable","message":"gateway closed"}',
          503,
        ),
      ),
    );

    await expectLater(
      client.fetchHealth(),
      throwsA(
        isA<ProviderApiException>()
            .having((error) => error.statusCode, 'statusCode', 503)
            .having((error) => error.message, 'message', 'gateway closed')
            .having(
              (error) => error.toString(),
              'toString',
              contains('HTTP 503'),
            ),
      ),
    );
  });

  test('ProviderApiClient falls back when error bodies are not JSON', () async {
    final client = ProviderApiClient(
      baseUrl: 'https://api.example.com',
      httpClient: MockClient(
        (_) async =>
            http.Response('not json', 500, reasonPhrase: 'bad gateway'),
      ),
    );

    await expectLater(
      client.fetchHealth(),
      throwsA(
        isA<ProviderApiException>()
            .having((error) => error.message, 'message', 'bad gateway'),
      ),
    );
  });
}

const _userWithoutNameBody = '''
{
  "user": {
    "id": "user-1",
    "email": "viewer@example.com",
    "role": "viewer",
    "status": "active"
  }
}
''';

const _healthBody = '''
{
  "service": "qtcloud-asset-provider",
  "status": "ok"
}
''';

const _adminUserBody = '''
{
  "user": {
    "id": "admin-1",
    "account": "admin",
    "email": "admin@example.com",
    "name": "Admin User",
    "role": "admin",
    "status": "active"
  }
}
''';

const _bucketsBody = '''
{
  "buckets": [
    {
      "name": "qtcloud-asset-studio",
      "region": "oss-cn-hangzhou",
      "storage_class": "Standard",
      "created_at": "2026-08-25"
    },
    {
      "name": "qtadmin-private",
      "region": "oss-cn-hangzhou",
      "storage_class": "Archive",
      "created_at": "2026-08-24"
    }
  ]
}
''';

const _objectsFirstPageBody = '''
{
  "objects": [
    {
      "key": "root.txt",
      "size": 12,
      "type": "Normal",
      "storage_class": "Standard",
      "last_modified": "2026-08-24 10:00:00"
    }
  ],
  "truncated": true,
  "next_marker": "next marker"
}
''';

const _objectsSecondPageBody = '''
{
  "objects": [
    {
      "key": "folder/file.txt",
      "size": 2048,
      "type": "Normal",
      "storage_class": "Standard",
      "last_modified": "2026-08-24 10:01:00"
    }
  ],
  "truncated": false,
  "next_marker": ""
}
''';

const _usersBody = '''
{
  "users": [
    {
      "id": "user-1",
      "email": "viewer@example.com",
      "name": "Viewer User",
      "role": "viewer",
      "status": "active"
    }
  ]
}
''';

const _invitedUserBody = '''
{
  "user": {
    "id": "new-1",
    "account": "new-user",
    "name": "New User",
    "role": "admin",
    "status": "active"
  }
}
''';

const _updatedUserBody = '''
{
  "user": {
    "id": "user-1",
    "email": "viewer@example.com",
    "name": "Viewer User",
    "role": "admin",
    "status": "active"
  }
}
''';

const _shareBody = '''
{
  "share": {
    "token": "share-token",
    "title": "公开资料",
    "bucket": "qtcloud-asset-studio",
    "prefixes": ["design/", "docs/"],
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
