// Basic widget smoke test for 量潮资产云 Studio.

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

import 'package:qtcloud_studio/main.dart';

void setTestViewSize(WidgetTester tester, Size size) {
  tester.view.devicePixelRatio = 1.0;
  tester.view.physicalSize = size;
  addTearDown(tester.view.resetPhysicalSize);
  addTearDown(tester.view.resetDevicePixelRatio);
}

void main() {
  test('provider URL defaults to the local development endpoint', () {
    const expectedProviderBaseUrl = String.fromEnvironment(
      'EXPECTED_PROVIDER_BASE_URL',
      defaultValue: 'http://127.0.0.1:9000',
    );
    expect(providerBaseUrl, expectedProviderBaseUrl);
  });

  testWidgets('Dashboard renders title and bucket section', (
    WidgetTester tester,
  ) async {
    await tester.pumpWidget(
      QtCloudAssetStudio(
        client: _authenticatedStudioClient(),
        loginRedirect: (_) {},
      ),
    );
    await tester.pumpAndSettle();

    // Verify the dashboard title is present.
    expect(find.text('量潮资产云'), findsOneWidget);

    // Verify the bucket section heading is present.
    expect(find.text('对象存储桶'), findsOneWidget);
  });

  testWidgets(
      'Dashboard redirects unauthenticated users before loading buckets', (
    WidgetTester tester,
  ) async {
    Uri? redirectedTo;
    final requestedPaths = <String>[];
    final client = ProviderApiClient(
      baseUrl: 'https://api.example.com',
      httpClient: MockClient((request) async {
        requestedPaths.add(request.url.path);
        if (request.url.path == '/auth/me') {
          return http.Response('{"error":"authentication required"}', 401);
        }
        return http.Response('{"error":"unexpected"}', 500);
      }),
    );

    await tester.pumpWidget(
      MaterialApp(
        home: DashboardScreen(
          client: client,
          loginRedirect: (uri) {
            redirectedTo = uri;
          },
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(requestedPaths, ['/auth/me']);
    expect(redirectedTo, Uri.parse('https://api.example.com/auth/login'));
    expect(find.text('正在跳转登录…'), findsOneWidget);
    expect(find.text('对象存储桶'), findsNothing);
  });

  testWidgets('Dashboard shows the authenticated user and logout action', (
    WidgetTester tester,
  ) async {
    final client = _authenticatedStudioClient();

    await tester.pumpWidget(
      MaterialApp(
        home: DashboardScreen(
          client: client,
          loginRedirect: (_) {},
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Viewer User'), findsOneWidget);
    expect(find.text('viewer'), findsOneWidget);
    expect(find.byTooltip('退出登录'), findsOneWidget);
    expect(find.text('对象存储桶'), findsOneWidget);
    expect(find.text('qtcloud-asset-studio'), findsOneWidget);
    expect(find.text('qtadmin-private'), findsNothing);
    expect(find.text('quanttide-terraform-state'), findsNothing);
  });

  testWidgets('Dashboard hides metadata-only buckets for viewers', (
    WidgetTester tester,
  ) async {
    await tester.pumpWidget(
      QtCloudAssetStudio(
        client: _authenticatedStudioClient(),
        loginRedirect: (_) {},
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('全部 1'), findsOneWidget);
    expect(find.text('qtcloud-asset-studio'), findsOneWidget);
    expect(find.text('qtadmin-private'), findsNothing);
    expect(find.text('quanttide-terraform-state'), findsNothing);
  });

  testWidgets('Dashboard keeps all buckets visible for admins', (
    WidgetTester tester,
  ) async {
    await tester.pumpWidget(
      QtCloudAssetStudio(
        client: _authenticatedStudioClient(userBody: _adminAuthMeBody),
        loginRedirect: (_) {},
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('全部 3'), findsOneWidget);
    expect(find.text('qtcloud-asset-studio'), findsOneWidget);
    expect(find.text('qtadmin-private'), findsOneWidget);
    expect(find.text('quanttide-terraform-state'), findsOneWidget);
  });

  testWidgets('Dashboard shows user management only for admins', (
    WidgetTester tester,
  ) async {
    await tester.pumpWidget(
      MaterialApp(
        home: DashboardScreen(
          client: _authenticatedStudioClient(userBody: _authMeBody),
          loginRedirect: (_) {},
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.byTooltip('用户管理'), findsNothing);

    await tester.pumpWidget(
      MaterialApp(
        home: DashboardScreen(
          client: _authenticatedStudioClient(userBody: _adminAuthMeBody),
          loginRedirect: (_) {},
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.byTooltip('用户管理'), findsOneWidget);
  });

  testWidgets('Admin user management lists users and submits invites', (
    WidgetTester tester,
  ) async {
    final requested = <String>[];
    final client = ProviderApiClient(
      baseUrl: 'https://api.example.com',
      httpClient: MockClient((request) async {
        requested.add('${request.method} ${request.url.path}');
        return switch ((request.method, request.url.path)) {
          ('GET', '/admin/users') => http.Response(_usersBody, 200),
          ('POST', '/admin/users') => http.Response(_invitedUserBody, 201),
          _ => http.Response('{"error":"unexpected"}', 500),
        };
      }),
    );

    await tester.pumpWidget(
      MaterialApp(
        home: AdminUsersScreen(client: client),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Admin User'), findsOneWidget);
    expect(find.text('Viewer User'), findsOneWidget);

    await tester.enterText(
        find.byKey(const Key('invite-email')), 'new@example.com');
    await tester.enterText(find.byKey(const Key('invite-name')), 'New User');
    await tester.tap(find.text('邀请'));
    await tester.pumpAndSettle();

    expect(requested, contains('POST /admin/users'));
    expect(find.text('邀请已创建'), findsOneWidget);
  });

  testWidgets('Logout revokes the server session and waits for manual login', (
    WidgetTester tester,
  ) async {
    Uri? redirectedTo;
    final requested = <String>[];
    final client = ProviderApiClient(
      baseUrl: 'https://api.example.com',
      httpClient: MockClient((request) async {
        requested.add('${request.method} ${request.url.path}');
        return switch ((request.method, request.url.path)) {
          ('GET', '/auth/me') => http.Response(_authMeBody, 200),
          ('GET', '/health') => http.Response(_healthBody, 200),
          ('GET', '/buckets') => http.Response(_bucketsBody, 200),
          ('POST', '/auth/logout') => http.Response('', 204),
          _ => http.Response('{"error":"unexpected"}', 500),
        };
      }),
    );

    await tester.pumpWidget(
      MaterialApp(
        home: DashboardScreen(
          client: client,
          loginRedirect: (uri) {
            redirectedTo = uri;
          },
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.byTooltip('退出登录'));
    await tester.pumpAndSettle();

    expect(requested, contains('POST /auth/logout'));
    expect(redirectedTo, isNull);
    expect(find.text('已退出当前会话，请重新登录。'), findsOneWidget);

    await tester.tap(find.text('重新登录'));
    await tester.pumpAndSettle();

    expect(redirectedTo, Uri.parse('https://api.example.com/auth/login'));
  });

  testWidgets('Dashboard places creation-time sort before alphabet sort', (
    WidgetTester tester,
  ) async {
    setTestViewSize(tester, const Size(1280, 720));

    await tester.pumpWidget(
      QtCloudAssetStudio(
        client: _authenticatedStudioClient(),
        loginRedirect: (_) {},
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('创建时间'), findsOneWidget);
    expect(find.text('A→Z'), findsOneWidget);

    final createdAtX = tester.getTopLeft(find.text('创建时间')).dx;
    final alphaX = tester.getTopLeft(find.text('A→Z')).dx;

    expect(createdAtX, lessThan(alphaX));

    await tester.tap(find.text('A→Z'));
    await tester.pump();
    expect(find.text('Z→A'), findsOneWidget);

    final createdAtButton = find.ancestor(
      of: find.text('创建时间'),
      matching: find.byType(TextButton),
    );

    expect(
      find.descendant(
        of: createdAtButton,
        matching: find.byIcon(Icons.arrow_downward),
      ),
      findsOneWidget,
    );

    await tester.tap(find.text('创建时间'));
    await tester.pump();

    expect(
      find.descendant(
        of: createdAtButton,
        matching: find.byIcon(Icons.arrow_upward),
      ),
      findsOneWidget,
    );
  });

  testWidgets('Dashboard exposes independent four-state bucket sorting', (
    WidgetTester tester,
  ) async {
    setTestViewSize(tester, const Size(1280, 720));

    await tester.pumpWidget(
      QtCloudAssetStudio(
        client: _authenticatedStudioClient(),
        loginRedirect: (_) {},
      ),
    );
    await tester.pumpAndSettle();

    final nameButton = find.ancestor(
      of: find.text('A→Z'),
      matching: find.byType(TextButton),
    );
    expect(
      find.descendant(
          of: nameButton, matching: find.byIcon(Icons.arrow_upward)),
      findsOneWidget,
    );

    await tester.tap(find.text('A→Z'));
    await tester.pump();
    expect(find.text('Z→A'), findsOneWidget);

    await tester.tap(find.text('Z→A'));
    await tester.pump();
    final disabledNameButton = find.ancestor(
      of: find.text('桶名：关闭'),
      matching: find.byType(TextButton),
    );
    expect(
      find.descendant(
        of: disabledNameButton,
        matching: find.byIcon(Icons.remove),
      ),
      findsOneWidget,
    );

    final createdAtButton = find.ancestor(
      of: find.text('创建时间'),
      matching: find.byType(TextButton),
    );
    expect(
      find.descendant(
        of: createdAtButton,
        matching: find.byIcon(Icons.arrow_downward),
      ),
      findsOneWidget,
    );

    await tester.tap(find.text('创建时间'));
    await tester.pump();
    expect(
      find.descendant(
        of: createdAtButton,
        matching: find.byIcon(Icons.arrow_upward),
      ),
      findsOneWidget,
    );

    await tester.tap(find.text('创建时间'));
    await tester.pump();
    final disabledCreatedAtButton = find.ancestor(
      of: find.text('创建时间：关闭'),
      matching: find.byType(TextButton),
    );
    expect(
      find.descendant(
        of: disabledCreatedAtButton,
        matching: find.byIcon(Icons.remove),
      ),
      findsOneWidget,
    );
  });

  testWidgets('bucket cards stay compact on wide screens', (
    WidgetTester tester,
  ) async {
    setTestViewSize(tester, const Size(1280, 720));

    const buckets = [
      Bucket(
        name: 'qtcloud-agent-site',
        region: 'cn-hangzhou',
        storageClass: 'Standard',
        createdAt: '2026-08-21',
      ),
      Bucket(
        name: 'qtcloud-agent-studio',
        region: 'cn-hangzhou',
        storageClass: 'Standard',
        createdAt: '2026-08-20',
      ),
      Bucket(
        name: 'qtbusiness-site',
        region: 'cn-hangzhou',
        storageClass: 'Standard',
        createdAt: '2026-08-19',
      ),
    ];

    await tester.pumpWidget(
      MaterialApp(
        home: SizedBox(
          width: 1280,
          height: 300,
          child: BucketListView(buckets: Future.value(buckets)),
        ),
      ),
    );
    await tester.pumpAndSettle();

    final card = tester.renderObject<RenderBox>(
      find.byType(BucketCard).first,
    );
    expect(card.size.width, lessThanOrEqualTo(280));
  });

  testWidgets('bucket grid shows six cards per row on wide screens', (
    WidgetTester tester,
  ) async {
    setTestViewSize(tester, const Size(1280, 720));

    final buckets = List.generate(
      7,
      (index) => Bucket(
        name: 'qtcloud-${index + 1}-site',
        region: 'cn-hangzhou',
        storageClass: 'Standard',
        createdAt: '2026-08-21',
      ),
    );

    await tester.pumpWidget(
      MaterialApp(
        home: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 1400),
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 24),
              child: SizedBox(
                height: 360,
                child: BucketListView(buckets: Future.value(buckets)),
              ),
            ),
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    final cards = find.byType(BucketCard);
    final firstRowTop = tester.getTopLeft(cards.at(0)).dy;

    for (var index = 1; index < 6; index += 1) {
      expect(tester.getTopLeft(cards.at(index)).dy, firstRowTop);
    }
    expect(tester.getTopLeft(cards.at(6)).dy, greaterThan(firstRowTop));
  });

  testWidgets('bucket grid keeps three rows per page on narrow screens', (
    WidgetTester tester,
  ) async {
    setTestViewSize(tester, const Size(620, 720));

    final buckets = List.generate(
      10,
      (index) {
        final suffix = (index + 1).toString().padLeft(2, '0');
        return Bucket(
          name: 'qtcloud-$suffix-site',
          region: 'cn-hangzhou',
          storageClass: 'Standard',
          createdAt: '2026-08-21',
        );
      },
    );

    await tester.pumpWidget(
      MaterialApp(
        home: SizedBox(
          width: 620,
          height: 500,
          child: BucketListView(buckets: Future.value(buckets)),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('qtcloud-09-site'), findsOneWidget);
    expect(find.text('qtcloud-10-site'), findsNothing);

    await tester.tap(find.byTooltip('下一页'));
    await tester.pumpAndSettle();

    expect(find.text('qtcloud-01-site'), findsNothing);
    expect(find.text('qtcloud-10-site'), findsOneWidget);
  });

  testWidgets('bucket cards sort by name alphabetically', (
    WidgetTester tester,
  ) async {
    const buckets = [
      Bucket(
        name: 'zeta-site',
        region: 'cn-hangzhou',
        storageClass: 'Standard',
        createdAt: '2026-08-19',
      ),
      Bucket(
        name: 'Alpha-site',
        region: 'cn-hangzhou',
        storageClass: 'Standard',
        createdAt: '2026-08-21',
      ),
      Bucket(
        name: 'beta-site',
        region: 'cn-hangzhou',
        storageClass: 'Standard',
        createdAt: '2026-08-20',
      ),
    ];

    await tester.pumpWidget(
      MaterialApp(
        home: SizedBox(
          width: 1280,
          height: 300,
          child: BucketListView(
            buckets: Future.value(buckets),
            sortMode: BucketSortMode.name,
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    final alphaX = tester.getTopLeft(find.text('Alpha-site')).dx;
    final betaX = tester.getTopLeft(find.text('beta-site')).dx;
    final zetaX = tester.getTopLeft(find.text('zeta-site')).dx;

    expect(alphaX, lessThan(betaX));
    expect(betaX, lessThan(zetaX));
  });

  testWidgets('bucket cards sort by creation time newest first', (
    WidgetTester tester,
  ) async {
    const buckets = [
      Bucket(
        name: 'older-site',
        region: 'cn-hangzhou',
        storageClass: 'Standard',
        createdAt: '2026-08-19',
      ),
      Bucket(
        name: 'newer-site',
        region: 'cn-hangzhou',
        storageClass: 'Standard',
        createdAt: '2026-08-21',
      ),
      Bucket(
        name: 'middle-site',
        region: 'cn-hangzhou',
        storageClass: 'Standard',
        createdAt: '2026-08-20',
      ),
    ];

    await tester.pumpWidget(
      MaterialApp(
        home: SizedBox(
          width: 1280,
          height: 300,
          child: BucketListView(
            buckets: Future.value(buckets),
            sortMode: BucketSortMode.createdAt,
            createdAtDescending: true,
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    final newerX = tester.getTopLeft(find.text('newer-site')).dx;
    final middleX = tester.getTopLeft(find.text('middle-site')).dx;
    final olderX = tester.getTopLeft(find.text('older-site')).dx;

    expect(newerX, lessThan(middleX));
    expect(middleX, lessThan(olderX));
  });

  testWidgets('bucket cards use name direction within the same creation date', (
    WidgetTester tester,
  ) async {
    const buckets = [
      Bucket(
        name: 'alpha-site',
        region: 'cn-hangzhou',
        storageClass: 'Standard',
        createdAt: '2026-08-21',
      ),
      Bucket(
        name: 'zeta-site',
        region: 'cn-hangzhou',
        storageClass: 'Standard',
        createdAt: '2026-08-21',
      ),
      Bucket(
        name: 'newer-site',
        region: 'cn-hangzhou',
        storageClass: 'Standard',
        createdAt: '2026-08-22',
      ),
    ];

    await tester.pumpWidget(
      MaterialApp(
        home: SizedBox(
          width: 1280,
          height: 300,
          child: BucketListView(
            buckets: Future.value(buckets),
            sortMode: BucketSortMode.createdAtThenName,
            sortAscending: false,
            createdAtDescending: true,
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    final newerX = tester.getTopLeft(find.text('newer-site')).dx;
    final zetaX = tester.getTopLeft(find.text('zeta-site')).dx;
    final alphaX = tester.getTopLeft(find.text('alpha-site')).dx;

    expect(newerX, lessThan(zetaX));
    expect(zetaX, lessThan(alphaX));
  });

  testWidgets('bucket cards update when creation-time direction changes', (
    WidgetTester tester,
  ) async {
    const buckets = [
      Bucket(
        name: 'older-site',
        region: 'cn-hangzhou',
        storageClass: 'Standard',
        createdAt: '2026-08-19',
      ),
      Bucket(
        name: 'newer-site',
        region: 'cn-hangzhou',
        storageClass: 'Standard',
        createdAt: '2026-08-21',
      ),
      Bucket(
        name: 'middle-site',
        region: 'cn-hangzhou',
        storageClass: 'Standard',
        createdAt: '2026-08-20',
      ),
    ];

    Widget buildList({required bool descending}) {
      return MaterialApp(
        home: SizedBox(
          width: 1280,
          height: 300,
          child: BucketListView(
            buckets: Future.value(buckets),
            sortMode: BucketSortMode.createdAt,
            createdAtDescending: descending,
          ),
        ),
      );
    }

    await tester.pumpWidget(buildList(descending: true));
    await tester.pumpAndSettle();

    expect(
      tester.getTopLeft(find.text('newer-site')).dx,
      lessThan(tester.getTopLeft(find.text('older-site')).dx),
    );

    await tester.pumpWidget(buildList(descending: false));
    await tester.pumpAndSettle();

    expect(
      tester.getTopLeft(find.text('older-site')).dx,
      lessThan(tester.getTopLeft(find.text('newer-site')).dx),
    );
  });

  testWidgets('bucket list returns to first page when sort mode changes', (
    WidgetTester tester,
  ) async {
    setTestViewSize(tester, const Size(620, 720));

    final buckets = List.generate(
      10,
      (index) {
        final suffix = (index + 1).toString().padLeft(2, '0');
        return Bucket(
          name: 'qtcloud-$suffix-site',
          region: 'cn-hangzhou',
          storageClass: 'Standard',
          createdAt: '2026-08-${(20 - index).toString().padLeft(2, '0')}',
        );
      },
    );

    Widget buildList({required BucketSortMode sortMode}) {
      return MaterialApp(
        home: SizedBox(
          width: 620,
          height: 500,
          child: BucketListView(
            buckets: Future.value(buckets),
            sortMode: sortMode,
          ),
        ),
      );
    }

    await tester.pumpWidget(buildList(sortMode: BucketSortMode.name));
    await tester.pumpAndSettle();

    await tester.tap(find.byTooltip('下一页'));
    await tester.pumpAndSettle();

    expect(find.text('qtcloud-10-site'), findsOneWidget);

    await tester.pumpWidget(buildList(sortMode: BucketSortMode.createdAt));
    await tester.pumpAndSettle();

    expect(find.text('1 / 2'), findsOneWidget);
    expect(find.text('qtcloud-01-site'), findsOneWidget);
    expect(find.text('qtcloud-10-site'), findsNothing);
  });

  testWidgets('bucket cards preserve provider order when sorting is disabled', (
    WidgetTester tester,
  ) async {
    const buckets = [
      Bucket(
        name: 'zeta-site',
        region: 'cn-hangzhou',
        storageClass: 'Standard',
        createdAt: '2026-08-19',
      ),
      Bucket(
        name: 'alpha-site',
        region: 'cn-hangzhou',
        storageClass: 'Standard',
        createdAt: '2026-08-21',
      ),
      Bucket(
        name: 'middle-site',
        region: 'cn-hangzhou',
        storageClass: 'Standard',
        createdAt: '2026-08-20',
      ),
    ];

    await tester.pumpWidget(
      MaterialApp(
        home: SizedBox(
          width: 1280,
          height: 300,
          child: BucketListView(
            buckets: Future.value(buckets),
            sortMode: BucketSortMode.none,
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(
      tester.getTopLeft(find.text('zeta-site')).dx,
      lessThan(tester.getTopLeft(find.text('alpha-site')).dx),
    );
    expect(
      tester.getTopLeft(find.text('alpha-site')).dx,
      lessThan(tester.getTopLeft(find.text('middle-site')).dx),
    );
  });

  testWidgets('private bucket cards do not open object browser', (
    WidgetTester tester,
  ) async {
    const buckets = [
      Bucket(
        name: 'qtadmin-private',
        region: 'cn-hangzhou',
        storageClass: 'Standard',
        createdAt: '2026-07-18',
      ),
    ];

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: BucketListView(buckets: Future.value(buckets)),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('qtadmin-private'));
    await tester.pumpAndSettle();

    expect(find.byType(BucketObjectsScreen), findsNothing);
    expect(find.text('私密桶仅展示元数据'), findsOneWidget);
  });

  test('metadata-only buckets cannot expose object links', () {
    expect(canExposeObjectLinks('qtadmin-private'), isFalse);
    expect(canExposeObjectLinks('quanttide-terraform-state'), isFalse);
    expect(canExposeObjectLinks('qtcloud-asset-studio'), isTrue);
  });

  test('object pagination follows every continuation marker', () async {
    final requestedMarkers = <String>[];
    final pages = <String, Map<String, dynamic>>{
      '': {
        'objects': [
          {
            'key': 'first.txt',
            'size': 1,
            'type': 'Normal',
            'storage_class': 'Standard',
            'last_modified': '2026-08-24 10:00:00',
          },
        ],
        'truncated': true,
        'next_marker': 'page-2',
      },
      'page-2': {
        'objects': [
          {
            'key': 'second.txt',
            'size': 2,
            'type': 'Normal',
            'storage_class': 'Standard',
            'last_modified': '2026-08-24 10:01:00',
          },
        ],
        'truncated': false,
        'next_marker': '',
      },
    };

    final objects = await fetchAllObjectPages((marker) async {
      requestedMarkers.add(marker);
      return pages[marker]!;
    });

    expect(requestedMarkers, ['', 'page-2']);
    expect(objects.map((object) => object.key), ['first.txt', 'second.txt']);
  });
}

ProviderApiClient _authenticatedStudioClient({String userBody = _authMeBody}) {
  return ProviderApiClient(
    baseUrl: 'https://api.example.com',
    httpClient: MockClient((request) async {
      return switch ((request.method, request.url.path)) {
        ('GET', '/auth/me') => http.Response(userBody, 200),
        ('GET', '/health') => http.Response(_healthBody, 200),
        ('GET', '/buckets') => http.Response(_bucketsBody, 200),
        _ => http.Response('{"error":"unexpected"}', 500),
      };
    }),
  );
}

const _authMeBody = '''
{
  "user": {
    "id": "user-1",
    "external_id": "lark-user-1",
    "email": "viewer@example.com",
    "name": "Viewer User",
    "role": "viewer",
    "status": "active",
    "created_at": "2026-08-25T00:00:00Z",
    "last_login_at": "2026-08-25T01:00:00Z"
  }
}
''';

const _adminAuthMeBody = '''
{
  "user": {
    "id": "admin-1",
    "external_id": "lark-admin-1",
    "email": "admin@example.com",
    "name": "Admin User",
    "role": "admin",
    "status": "active",
    "created_at": "2026-08-25T00:00:00Z",
    "last_login_at": "2026-08-25T01:00:00Z"
  }
}
''';

const _healthBody = '''
{
  "service": "qtcloud-asset-provider",
  "status": "ok"
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
      "storage_class": "Standard",
      "created_at": "2026-08-24"
    },
    {
      "name": "quanttide-terraform-state",
      "region": "oss-cn-hangzhou",
      "storage_class": "Standard",
      "created_at": "2026-08-23"
    }
  ]
}
''';

const _usersBody = '''
{
  "total": 2,
  "users": [
    {
      "id": "admin-1",
      "external_id": "lark-admin-1",
      "email": "admin@example.com",
      "name": "Admin User",
      "role": "admin",
      "status": "active",
      "created_at": "2026-08-25T00:00:00Z",
      "last_login_at": "2026-08-25T01:00:00Z"
    },
    {
      "id": "viewer-1",
      "external_id": "lark-viewer-1",
      "email": "viewer@example.com",
      "name": "Viewer User",
      "role": "viewer",
      "status": "active",
      "created_at": "2026-08-25T00:00:00Z",
      "last_login_at": "2026-08-25T01:00:00Z"
    }
  ]
}
''';

const _invitedUserBody = '''
{
  "user": {
    "id": "new-1",
    "external_id": "new@example.com",
    "email": "new@example.com",
    "name": "New User",
    "role": "viewer",
    "status": "active",
    "created_at": "2026-08-25T00:00:00Z",
    "last_login_at": "0001-01-01T00:00:00Z"
  }
}
''';
