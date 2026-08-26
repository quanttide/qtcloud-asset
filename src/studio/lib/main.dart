import 'dart:convert';
import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:http/http.dart' as http;

import 'login_redirect.dart';
import 'provider_http_client.dart';

const providerBaseUrl = String.fromEnvironment(
  'PROVIDER_BASE_URL',
  defaultValue: 'http://127.0.0.1:9000',
);

const _bucketGridMaxColumns = 6;
const _bucketGridRowsPerPage = 3;
const _bucketGridMinCardWidth = 180.0;
const _bucketGridSpacing = 8.0;
const _bucketGridMainExtent = 104.0;

enum BucketSortMode { none, name, createdAt, createdAtThenName }

class AuthRequiredException implements Exception {
  const AuthRequiredException();

  @override
  String toString() => 'authentication required';
}

class ProviderApiException implements Exception {
  const ProviderApiException(this.statusCode, this.message);

  final int statusCode;
  final String message;

  @override
  String toString() => 'Provider returned HTTP $statusCode: $message';
}

class ProviderUser {
  const ProviderUser({
    required this.id,
    required this.email,
    required this.name,
    required this.role,
    required this.status,
  });

  final String id;
  final String email;
  final String name;
  final String role;
  final String status;

  factory ProviderUser.fromJson(Map<String, dynamic> json) {
    return ProviderUser(
      id: json['id'] as String? ?? '',
      email: json['email'] as String? ?? '',
      name: json['name'] as String? ?? '',
      role: json['role'] as String? ?? '',
      status: json['status'] as String? ?? '',
    );
  }

  String get displayName => name.isNotEmpty ? name : email;
}

class ProviderApiClient {
  ProviderApiClient({String baseUrl = providerBaseUrl, http.Client? httpClient})
      : baseUrl = baseUrl.replaceFirst(RegExp(r'/+$'), ''),
        _httpClient = httpClient ?? createProviderHttpClient();

  final String baseUrl;
  final http.Client _httpClient;

  Uri get loginUri => Uri.parse('$baseUrl/auth/login');

  Future<ProviderUser> login({
    required String email,
    required String password,
  }) async {
    final response = await _httpClient
        .post(
          _uri('/auth/login'),
          headers: {'Content-Type': 'application/json'},
          body: jsonEncode({'email': email, 'password': password}),
        )
        .timeout(const Duration(seconds: 15));
    _ensureSuccess(response);
    final body = jsonDecode(response.body) as Map<String, dynamic>;
    return ProviderUser.fromJson(body['user'] as Map<String, dynamic>);
  }

  Uri _uri(String path, {Map<String, String>? queryParameters}) {
    return Uri.parse('$baseUrl$path').replace(
      queryParameters: queryParameters == null || queryParameters.isEmpty
          ? null
          : queryParameters,
    );
  }

  Future<ProviderUser> fetchCurrentUser() async {
    final response = await _httpClient
        .get(_uri('/auth/me'))
        .timeout(const Duration(seconds: 15));
    if (response.statusCode == 401) {
      throw const AuthRequiredException();
    }
    _ensureSuccess(response);
    final body = jsonDecode(response.body) as Map<String, dynamic>;
    final user = body['user'] as Map<String, dynamic>;
    return ProviderUser.fromJson(user);
  }

  Future<void> logout() async {
    final response = await _httpClient
        .post(_uri('/auth/logout'))
        .timeout(const Duration(seconds: 15));
    if (response.statusCode == 401) return;
    _ensureSuccess(response, allowedStatusCodes: {204});
  }

  Future<Map<String, dynamic>> fetchHealth() async {
    final response = await _httpClient
        .get(_uri('/health'))
        .timeout(const Duration(seconds: 15));
    _ensureSuccess(response);
    return jsonDecode(response.body) as Map<String, dynamic>;
  }

  Future<List<Bucket>> fetchBuckets() async {
    final response = await _httpClient
        .get(_uri('/buckets'))
        .timeout(const Duration(seconds: 15));
    if (response.statusCode == 401) {
      throw const AuthRequiredException();
    }
    _ensureSuccess(response);
    final body = jsonDecode(response.body) as Map<String, dynamic>;
    final list = body['buckets'] as List<dynamic>;
    return list.map((e) => Bucket.fromJson(e as Map<String, dynamic>)).toList();
  }

  Future<List<OssObject>> fetchObjects(String bucketName) async {
    final name = Uri.encodeComponent(bucketName);
    return fetchAllObjectPages((marker) async {
      final query = marker.isEmpty ? null : {'marker': marker};
      final response = await _httpClient
          .get(_uri('/buckets/$name/objects', queryParameters: query))
          .timeout(const Duration(seconds: 20));
      if (response.statusCode == 401) {
        throw const AuthRequiredException();
      }
      _ensureSuccess(response);
      return jsonDecode(response.body) as Map<String, dynamic>;
    });
  }

  Future<String> fetchObjectUrl({
    required String bucketName,
    required String objectKey,
    required int expiresSeconds,
  }) async {
    final bucket = Uri.encodeComponent(bucketName);
    final response = await _httpClient
        .get(
          _uri(
            '/buckets/$bucket/object-url',
            queryParameters: {
              'key': objectKey,
              'expires': '$expiresSeconds',
            },
          ),
        )
        .timeout(const Duration(seconds: 20));
    if (response.statusCode == 401) {
      throw const AuthRequiredException();
    }
    _ensureSuccess(response);
    final body = jsonDecode(response.body) as Map<String, dynamic>;
    return body['url'] as String;
  }

  Future<List<ProviderUser>> fetchUsers() async {
    final response = await _httpClient
        .get(_uri('/admin/users'))
        .timeout(const Duration(seconds: 15));
    if (response.statusCode == 401) {
      throw const AuthRequiredException();
    }
    _ensureSuccess(response);
    final body = jsonDecode(response.body) as Map<String, dynamic>;
    final list = body['users'] as List<dynamic>;
    return list
        .map((item) => ProviderUser.fromJson(item as Map<String, dynamic>))
        .toList();
  }

  Future<ProviderUser> inviteUser({
    required String email,
    required String name,
    String role = 'viewer',
  }) async {
    final response = await _httpClient
        .post(
          _uri('/admin/users'),
          headers: {'Content-Type': 'application/json'},
          body: jsonEncode({
            'email': email,
            'name': name,
            'role': role,
          }),
        )
        .timeout(const Duration(seconds: 15));
    if (response.statusCode == 401) {
      throw const AuthRequiredException();
    }
    _ensureSuccess(response, allowedStatusCodes: {201});
    final body = jsonDecode(response.body) as Map<String, dynamic>;
    return ProviderUser.fromJson(body['user'] as Map<String, dynamic>);
  }

  Future<ProviderUser> updateUserRole({
    required String userId,
    required String role,
  }) async {
    final response = await _httpClient
        .patch(
          _uri('/admin/users/$userId/role'),
          headers: {'Content-Type': 'application/json'},
          body: jsonEncode({'role': role}),
        )
        .timeout(const Duration(seconds: 15));
    if (response.statusCode == 401) {
      throw const AuthRequiredException();
    }
    _ensureSuccess(response);
    final body = jsonDecode(response.body) as Map<String, dynamic>;
    return ProviderUser.fromJson(body['user'] as Map<String, dynamic>);
  }

  Future<void> disableUser(String userId) async {
    final response = await _httpClient
        .post(_uri('/admin/users/$userId/disable'))
        .timeout(const Duration(seconds: 15));
    if (response.statusCode == 401) {
      throw const AuthRequiredException();
    }
    _ensureSuccess(response);
  }

  Future<void> revokeUserSessions(String userId) async {
    final response = await _httpClient
        .post(_uri('/admin/users/$userId/sessions/revoke'))
        .timeout(const Duration(seconds: 15));
    if (response.statusCode == 401) {
      throw const AuthRequiredException();
    }
    _ensureSuccess(response);
  }

  void _ensureSuccess(
    http.Response response, {
    Set<int> allowedStatusCodes = const {},
  }) {
    if (response.statusCode >= 200 && response.statusCode < 300 ||
        allowedStatusCodes.contains(response.statusCode)) {
      return;
    }
    throw ProviderApiException(response.statusCode, _errorMessage(response));
  }

  String _errorMessage(http.Response response) {
    try {
      final body = jsonDecode(response.body) as Map<String, dynamic>;
      return body['error']?.toString() ??
          response.reasonPhrase ??
          'request failed';
    } catch (_) {
      return response.reasonPhrase ?? 'request failed';
    }
  }
}

bool canExposeObjectLinks(String bucketName) {
  return !isMetadataOnlyBucket(bucketName);
}

bool isMetadataOnlyBucket(String bucketName) {
  return bucketName.endsWith('-private') ||
      bucketName == 'quanttide-terraform-state';
}

List<Bucket> visibleBucketsForRole(String role, Iterable<Bucket> buckets) {
  if (role == 'admin') return List<Bucket>.of(buckets);
  return buckets.where((bucket) => !isMetadataOnlyBucket(bucket.name)).toList();
}

typedef ObjectPageLoader = Future<Map<String, dynamic>> Function(String marker);

Future<List<OssObject>> fetchAllObjectPages(ObjectPageLoader loadPage) async {
  final objects = <OssObject>[];
  var marker = '';

  while (true) {
    final body = await loadPage(marker);
    final page = body['objects'] as List<dynamic>? ?? const [];
    objects.addAll(
      page.map((item) => OssObject.fromJson(item as Map<String, dynamic>)),
    );

    final truncated = body['truncated'] == true;
    final nextMarker = body['next_marker'] as String? ?? '';
    if (!truncated || nextMarker.isEmpty || nextMarker == marker) {
      return objects;
    }
    marker = nextMarker;
  }
}

int _bucketGridColumnCount(double width) {
  if (!width.isFinite) {
    return _bucketGridMaxColumns;
  }
  if (width <= 0) {
    return 1;
  }

  final columns = ((width + _bucketGridSpacing) /
          (_bucketGridMinCardWidth + _bucketGridSpacing))
      .floor();
  return math.max(1, math.min(_bucketGridMaxColumns, columns));
}

void main() {
  runApp(const QtCloudAssetStudio());
}

class QtCloudAssetStudio extends StatelessWidget {
  const QtCloudAssetStudio({super.key, this.client, this.loginRedirect});

  final ProviderApiClient? client;
  final LoginRedirect? loginRedirect;

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: '量潮资产云',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: const Color(0xFF2563EB)),
        useMaterial3: true,
      ),
      home: DashboardScreen(
        client: client,
        loginRedirect: loginRedirect,
      ),
    );
  }
}

class Bucket {
  const Bucket({
    required this.name,
    required this.region,
    required this.storageClass,
    required this.createdAt,
  });

  final String name;
  final String region;
  final String storageClass;
  final String createdAt;

  factory Bucket.fromJson(Map<String, dynamic> json) {
    return Bucket(
      name: json['name'] as String,
      region: json['region'] as String? ?? '',
      storageClass: json['storage_class'] as String? ?? '',
      createdAt: json['created_at'] as String? ?? '',
    );
  }

  String get category {
    if (name.endsWith('-studio')) return 'Studio';
    if (name.endsWith('-private')) return 'Private';
    if (name.endsWith('-site')) return 'Site';
    if (name.endsWith('-provider')) return 'Provider';
    return 'Other';
  }

  bool get canBrowseObjects => canExposeObjectLinks(name);

  /// 每个用途分类的专属图标。
  IconData get categoryIcon {
    switch (category) {
      case 'Studio':
        return Icons.web;
      case 'Private':
        return Icons.lock;
      case 'Site':
        return Icons.public;
      case 'Provider':
        return Icons.dns;
      default:
        return Icons.storage;
    }
  }

  /// 每个用途分类的主题色。
  Color get categoryColor {
    switch (category) {
      case 'Studio':
        return const Color(0xFF2563EB); // 蓝
      case 'Private':
        return const Color(0xFFEF4444); // 红
      case 'Site':
        return const Color(0xFF10B981); // 绿
      case 'Provider':
        return const Color(0xFF8B5CF6); // 紫
      default:
        return const Color(0xFF64748B); // 灰
    }
  }

  /// 基于桶名的稳定派生色，让同分类的桶之间也有颜色差异。
  Color get accentColor {
    final hash = name.codeUnits.fold<int>(0, (sum, c) => sum + c);
    final hue = (hash * 37) % 360;
    return HSLColor.fromAHSL(1, hue.toDouble(), 0.6, 0.55).toColor();
  }
}

class OssObject {
  const OssObject({
    required this.key,
    required this.size,
    required this.type,
    required this.storageClass,
    required this.lastModified,
  });

  final String key;
  final int size;
  final String type;
  final String storageClass;
  final String lastModified;

  factory OssObject.fromJson(Map<String, dynamic> json) {
    return OssObject(
      key: json['key'] as String,
      size: (json['size'] as num?)?.toInt() ?? 0,
      type: json['type'] as String? ?? '',
      storageClass: json['storage_class'] as String? ?? '',
      lastModified: json['last_modified'] as String? ?? '',
    );
  }

  /// 是否看起来像目录（以 / 结尾）
  bool get isDir => key.endsWith('/');

  /// 人类可读的文件大小
  String get sizeLabel {
    if (size >= 1024 * 1024 * 1024) {
      return '${(size / (1024 * 1024 * 1024)).toStringAsFixed(2)} GB';
    }
    if (size >= 1024 * 1024) {
      return '${(size / (1024 * 1024)).toStringAsFixed(2)} MB';
    }
    if (size >= 1024) {
      return '${(size / 1024).toStringAsFixed(2)} KB';
    }
    return '$size B';
  }
}

class DashboardScreen extends StatefulWidget {
  const DashboardScreen({super.key, this.client, this.loginRedirect});

  final ProviderApiClient? client;
  final LoginRedirect? loginRedirect;

  @override
  State<DashboardScreen> createState() => _DashboardScreenState();
}

class _DashboardScreenState extends State<DashboardScreen> {
  late ProviderApiClient _client;
  late Future<ProviderUser> _currentUser;
  Future<Map<String, dynamic>>? _health;
  Future<List<Bucket>>? _buckets;
  bool _loggedOut = false;
  bool _loginSubmitting = false;
  String? _loginError;
  final _loginEmailController = TextEditingController();
  final _loginPasswordController = TextEditingController();
  String? _selectedCategory; // null = 全部
  String _searchText = '';
  bool _sortAscending = true; // true = A 到 Z（默认），false = Z 到 A
  bool _createdAtDescending = true; // true = 新到旧（默认），false = 旧到新
  bool _nameSortEnabled = true;
  bool _createdAtSortEnabled = true;

  static const _categories = ['Studio', 'Private', 'Site', 'Provider'];

  BucketSortMode get _bucketSortMode {
    if (_createdAtSortEnabled && _nameSortEnabled) {
      return BucketSortMode.createdAtThenName;
    }
    if (_createdAtSortEnabled) return BucketSortMode.createdAt;
    if (_nameSortEnabled) return BucketSortMode.name;
    return BucketSortMode.none;
  }

  @override
  void initState() {
    super.initState();
    _client = widget.client ?? ProviderApiClient();
    _currentUser = _client.fetchCurrentUser();
  }

  @override
  void dispose() {
    _loginEmailController.dispose();
    _loginPasswordController.dispose();
    super.dispose();
  }

  @override
  void didUpdateWidget(covariant DashboardScreen oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.client != widget.client ||
        oldWidget.loginRedirect != widget.loginRedirect) {
      _client = widget.client ?? ProviderApiClient();
      _loggedOut = false;
      _loginError = null;
      _currentUser = _client.fetchCurrentUser();
      _health = null;
      _buckets = null;
    }
  }

  Future<Map<String, dynamic>> _fetchHealth() async {
    return _client.fetchHealth();
  }

  Future<List<Bucket>> _fetchBuckets() async {
    return _client.fetchBuckets();
  }

  void _refresh() {
    setState(() {
      _loggedOut = false;
      _loginError = null;
      _currentUser = _client.fetchCurrentUser();
      _health = null;
      _buckets = null;
    });
  }

  Future<void> _logout() async {
    try {
      await _client.logout();
    } finally {
      if (mounted) {
        setState(() {
          _loggedOut = true;
          _loginError = null;
          _health = null;
          _buckets = null;
        });
      }
    }
  }

  Future<void> _login() async {
    final email = _loginEmailController.text.trim();
    final password = _loginPasswordController.text;
    if (email.isEmpty || password.isEmpty) {
      setState(() {
        _loginError = '请输入账号和密码。';
      });
      return;
    }

    setState(() {
      _loginSubmitting = true;
      _loginError = null;
    });
    try {
      final user = await _client.login(email: email, password: password);
      if (!mounted) return;
      setState(() {
        _loggedOut = false;
        _loginSubmitting = false;
        _loginError = null;
        _currentUser = Future.value(user);
        _health = null;
        _buckets = null;
      });
    } catch (error) {
      if (!mounted) return;
      setState(() {
        _loginSubmitting = false;
        _loginError = error is ProviderApiException && error.statusCode == 503
            ? '登录服务尚未配置。'
            : '账号或密码错误，请重试。';
      });
    }
  }

  void _ensureProtectedDataLoaded() {
    _health ??= _fetchHealth();
    _buckets ??= _fetchBuckets();
  }

  Widget _buildSearchField() {
    return TextField(
      onChanged: (value) {
        setState(() {
          _searchText = value.trim();
        });
      },
      decoration: InputDecoration(
        hintText: '搜索桶名称…',
        prefixIcon: const Icon(Icons.search, size: 20),
        isDense: true,
        contentPadding: const EdgeInsets.symmetric(
          horizontal: 12,
          vertical: 10,
        ),
        filled: true,
        fillColor: Colors.white,
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
          borderSide: BorderSide(color: Colors.grey.shade300),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
          borderSide: BorderSide(color: Colors.grey.shade300),
        ),
      ),
    );
  }

  ButtonStyle _sortButtonStyle(bool active) {
    return TextButton.styleFrom(
      foregroundColor:
          active ? const Color(0xFF2563EB) : const Color(0xFF6B7280),
      backgroundColor:
          active ? const Color(0xFF2563EB).withValues(alpha: 0.08) : null,
    );
  }

  Widget _buildSortControls() {
    final createdAtLabel = _createdAtSortEnabled ? '创建时间' : '创建时间：关闭';
    final nameLabel = !_nameSortEnabled
        ? '桶名：关闭'
        : _sortAscending
            ? 'A→Z'
            : 'Z→A';

    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        TextButton.icon(
          onPressed: () {
            setState(() {
              if (!_createdAtSortEnabled) {
                _createdAtSortEnabled = true;
                _createdAtDescending = true;
              } else if (_createdAtDescending) {
                _createdAtDescending = !_createdAtDescending;
              } else {
                _createdAtSortEnabled = false;
              }
            });
          },
          icon: Icon(
            !_createdAtSortEnabled
                ? Icons.remove
                : _createdAtDescending
                    ? Icons.arrow_downward
                    : Icons.arrow_upward,
            size: 18,
          ),
          label: Text(createdAtLabel),
          style: _sortButtonStyle(_createdAtSortEnabled),
        ),
        const SizedBox(width: 8),
        TextButton.icon(
          onPressed: () {
            setState(() {
              if (!_nameSortEnabled) {
                _nameSortEnabled = true;
                _sortAscending = true;
              } else if (_sortAscending) {
                _sortAscending = false;
              } else {
                _nameSortEnabled = false;
              }
            });
          },
          icon: Icon(
            !_nameSortEnabled
                ? Icons.remove
                : _sortAscending
                    ? Icons.arrow_upward
                    : Icons.arrow_downward,
            size: 18,
          ),
          label: Text(nameLabel),
          style: _sortButtonStyle(_nameSortEnabled),
        ),
      ],
    );
  }

  Widget _buildSearchAndSortBar() {
    return LayoutBuilder(
      builder: (context, constraints) {
        final sortControls = _buildSortControls();

        if (constraints.maxWidth < 560) {
          return Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              _buildSearchField(),
              const SizedBox(height: 8),
              Align(
                alignment: Alignment.centerRight,
                child: sortControls,
              ),
            ],
          );
        }

        return Row(
          children: [
            Expanded(child: _buildSearchField()),
            const SizedBox(width: 8),
            sortControls,
          ],
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    final logoutMessage = _loggedOut ? '已退出当前会话，请重新登录。' : null;
    if (_loggedOut) {
      return _AuthScaffold(
        title: '量潮资产云',
        child: _LoginPanel(
          emailController: _loginEmailController,
          passwordController: _loginPasswordController,
          message: logoutMessage,
          error: _loginError,
          submitting: _loginSubmitting,
          onSubmit: _login,
        ),
      );
    }

    return FutureBuilder<ProviderUser>(
      future: _currentUser,
      builder: (context, snapshot) {
        if (snapshot.connectionState != ConnectionState.done) {
          return const _AuthScaffold(
            title: '量潮资产云',
            child: _AuthStatusPanel(
              icon: Icons.hourglass_empty,
              title: '正在确认登录状态…',
              message: '正在连接 Provider 会话。',
            ),
          );
        }

        if (snapshot.hasError) {
          final error = snapshot.error;
          if (error is AuthRequiredException) {
            return _AuthScaffold(
              title: '量潮资产云',
              child: _LoginPanel(
                emailController: _loginEmailController,
                passwordController: _loginPasswordController,
                error: _loginError,
                submitting: _loginSubmitting,
                onSubmit: _login,
              ),
            );
          }

          final forbidden =
              error is ProviderApiException && error.statusCode == 403;
          return _AuthScaffold(
            title: '量潮资产云',
            child: _AuthStatusPanel(
              icon: forbidden ? Icons.block : Icons.error_outline,
              title: forbidden ? '账号暂不可用' : '登录状态校验失败',
              message: error.toString(),
              actionLabel: '重试',
              onAction: _refresh,
            ),
          );
        }

        final user = snapshot.data!;
        final showMetadataOnlyBuckets = user.role == 'admin';
        _ensureProtectedDataLoaded();
        return Scaffold(
          backgroundColor: const Color(0xFFF7F8FA),
          appBar: AppBar(
            title: const Text('量潮资产云'),
            actions: [
              _CurrentUserBadge(user: user),
              if (user.role == 'admin')
                IconButton(
                  tooltip: '用户管理',
                  onPressed: () {
                    Navigator.of(context).push(
                      MaterialPageRoute(
                        builder: (_) => AdminUsersScreen(client: _client),
                      ),
                    );
                  },
                  icon: const Icon(Icons.manage_accounts),
                ),
              IconButton(
                tooltip: '退出登录',
                onPressed: _logout,
                icon: const Icon(Icons.logout),
              ),
              IconButton(
                tooltip: 'Refresh',
                onPressed: _refresh,
                icon: const Icon(Icons.refresh),
              ),
            ],
          ),
          body: Center(
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 1400),
              child: Padding(
                padding:
                    const EdgeInsets.symmetric(horizontal: 24, vertical: 24),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text(
                      '对象存储桶',
                      style:
                          TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
                    ),
                    const SizedBox(height: 12),
                    _CategoryFilterBar(
                      categories: _categories,
                      buckets: _buckets,
                      showMetadataOnlyBuckets: showMetadataOnlyBuckets,
                      selected: _selectedCategory,
                      onSelected: (value) {
                        setState(() {
                          _selectedCategory = value;
                        });
                      },
                    ),
                    const SizedBox(height: 12),
                    _buildSearchAndSortBar(),
                    const SizedBox(height: 12),
                    Expanded(
                      child: BucketListView(
                        buckets: _buckets,
                        client: _client,
                        showMetadataOnlyBuckets: showMetadataOnlyBuckets,
                        category: _selectedCategory,
                        searchText: _searchText,
                        sortMode: _bucketSortMode,
                        sortAscending: _sortAscending,
                        createdAtDescending: _createdAtDescending,
                      ),
                    ),
                    const SizedBox(height: 16),
                    _ProviderStatusCard(health: _health),
                  ],
                ),
              ),
            ),
          ),
        );
      },
    );
  }
}

class _AuthScaffold extends StatelessWidget {
  const _AuthScaffold({required this.title, required this.child});

  final String title;
  final Widget child;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF7F8FA),
      appBar: AppBar(title: Text(title)),
      body: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 420),
          child: child,
        ),
      ),
    );
  }
}

class _AuthStatusPanel extends StatelessWidget {
  const _AuthStatusPanel({
    required this.icon,
    required this.title,
    required this.message,
    this.actionLabel,
    this.onAction,
  });

  final IconData icon;
  final String title;
  final String message;
  final String? actionLabel;
  final VoidCallback? onAction;

  @override
  Widget build(BuildContext context) {
    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(icon, size: 32, color: const Color(0xFF2563EB)),
            const SizedBox(height: 12),
            Text(
              title,
              style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 8),
            Text(
              message,
              style: const TextStyle(fontSize: 13, color: Color(0xFF6B7280)),
              textAlign: TextAlign.center,
            ),
            if (actionLabel != null && onAction != null) ...[
              const SizedBox(height: 16),
              FilledButton.icon(
                onPressed: onAction,
                icon: const Icon(Icons.login, size: 18),
                label: Text(actionLabel!),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

class _LoginPanel extends StatelessWidget {
  const _LoginPanel({
    required this.emailController,
    required this.passwordController,
    required this.submitting,
    required this.onSubmit,
    this.message,
    this.error,
  });

  final TextEditingController emailController;
  final TextEditingController passwordController;
  final bool submitting;
  final VoidCallback onSubmit;
  final String? message;
  final String? error;

  @override
  Widget build(BuildContext context) {
    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const Icon(Icons.lock_person, size: 32, color: Color(0xFF2563EB)),
            const SizedBox(height: 12),
            const Text(
              '登录资产云',
              style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
              textAlign: TextAlign.center,
            ),
            if (message != null) ...[
              const SizedBox(height: 8),
              Text(
                message!,
                style: const TextStyle(fontSize: 13, color: Color(0xFF6B7280)),
                textAlign: TextAlign.center,
              ),
            ],
            const SizedBox(height: 18),
            TextField(
              key: const Key('login-email'),
              controller: emailController,
              enabled: !submitting,
              keyboardType: TextInputType.emailAddress,
              autofillHints: const [AutofillHints.username],
              decoration: const InputDecoration(
                labelText: '账号',
                prefixIcon: Icon(Icons.account_circle_outlined),
                isDense: true,
              ),
            ),
            const SizedBox(height: 12),
            TextField(
              key: const Key('login-password'),
              controller: passwordController,
              enabled: !submitting,
              obscureText: true,
              autofillHints: const [AutofillHints.password],
              onSubmitted: (_) => submitting ? null : onSubmit(),
              decoration: const InputDecoration(
                labelText: '密码',
                prefixIcon: Icon(Icons.password),
                isDense: true,
              ),
            ),
            if (error != null) ...[
              const SizedBox(height: 10),
              Text(
                error!,
                style: TextStyle(fontSize: 13, color: Colors.red.shade700),
              ),
            ],
            const SizedBox(height: 18),
            FilledButton.icon(
              onPressed: submitting ? null : onSubmit,
              icon: submitting
                  ? const SizedBox(
                      width: 16,
                      height: 16,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : const Icon(Icons.login, size: 18),
              label: const Text('登录'),
            ),
          ],
        ),
      ),
    );
  }
}

class _CurrentUserBadge extends StatelessWidget {
  const _CurrentUserBadge({required this.user});

  final ProviderUser user;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 8),
      child: Center(
        child: Container(
          constraints: const BoxConstraints(maxWidth: 240),
          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
          decoration: BoxDecoration(
            color: const Color(0xFF2563EB).withValues(alpha: 0.08),
            borderRadius: BorderRadius.circular(18),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Icon(Icons.account_circle, size: 18),
              const SizedBox(width: 6),
              Flexible(
                child: Text(
                  user.displayName,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(fontSize: 13),
                ),
              ),
              const SizedBox(width: 8),
              Text(
                user.role,
                style: const TextStyle(
                  fontSize: 12,
                  color: Color(0xFF2563EB),
                  fontWeight: FontWeight.w600,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class AdminUsersScreen extends StatefulWidget {
  const AdminUsersScreen({super.key, required this.client});

  final ProviderApiClient client;

  @override
  State<AdminUsersScreen> createState() => _AdminUsersScreenState();
}

class _AdminUsersScreenState extends State<AdminUsersScreen> {
  final _emailController = TextEditingController();
  final _nameController = TextEditingController();
  late Future<List<ProviderUser>> _users;
  String _inviteRole = 'viewer';
  bool _submitting = false;

  @override
  void initState() {
    super.initState();
    _users = widget.client.fetchUsers();
  }

  @override
  void dispose() {
    _emailController.dispose();
    _nameController.dispose();
    super.dispose();
  }

  Future<void> _refreshUsers() async {
    setState(() {
      _users = widget.client.fetchUsers();
    });
  }

  Future<void> _inviteUser() async {
    final email = _emailController.text.trim();
    final name = _nameController.text.trim();
    if (email.isEmpty || name.isEmpty) {
      _showMessage('请填写邮箱和姓名');
      return;
    }

    setState(() {
      _submitting = true;
    });
    try {
      await widget.client.inviteUser(
        email: email,
        name: name,
        role: _inviteRole,
      );
      _emailController.clear();
      _nameController.clear();
      await _refreshUsers();
      _showMessage('邀请已创建');
    } catch (error) {
      _showMessage('邀请失败：$error');
    } finally {
      if (mounted) {
        setState(() {
          _submitting = false;
        });
      }
    }
  }

  void _showMessage(String message) {
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(message)),
    );
  }

  Future<void> _updateRole(ProviderUser user, String role) async {
    try {
      await widget.client.updateUserRole(userId: user.id, role: role);
      await _refreshUsers();
      _showMessage('角色已更新');
    } catch (error) {
      _showMessage('角色更新失败：$error');
    }
  }

  Future<void> _disableUser(ProviderUser user) async {
    try {
      await widget.client.disableUser(user.id);
      await _refreshUsers();
      _showMessage('账号已停用');
    } catch (error) {
      _showMessage('停用失败：$error');
    }
  }

  Future<void> _revokeSessions(ProviderUser user) async {
    try {
      await widget.client.revokeUserSessions(user.id);
      _showMessage('会话已撤销');
    } catch (error) {
      _showMessage('撤销失败：$error');
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF7F8FA),
      appBar: AppBar(
        title: const Text('用户管理'),
        actions: [
          IconButton(
            tooltip: '刷新用户',
            onPressed: _refreshUsers,
            icon: const Icon(Icons.refresh),
          ),
        ],
      ),
      body: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 960),
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _InviteUserCard(
                  emailController: _emailController,
                  nameController: _nameController,
                  role: _inviteRole,
                  submitting: _submitting,
                  onRoleChanged: (value) {
                    if (value == null) return;
                    setState(() {
                      _inviteRole = value;
                    });
                  },
                  onSubmit: _inviteUser,
                ),
                const SizedBox(height: 16),
                Expanded(
                  child: FutureBuilder<List<ProviderUser>>(
                    future: _users,
                    builder: (context, snapshot) {
                      if (snapshot.connectionState != ConnectionState.done) {
                        return const Center(child: CircularProgressIndicator());
                      }
                      if (snapshot.hasError) {
                        return Center(
                          child: Text(
                            '加载用户失败：${snapshot.error}',
                            style: TextStyle(color: Colors.red.shade700),
                          ),
                        );
                      }
                      final users = snapshot.data ?? const [];
                      if (users.isEmpty) {
                        return const Center(child: Text('暂无用户'));
                      }
                      return ListView.separated(
                        itemCount: users.length,
                        separatorBuilder: (_, __) => const SizedBox(height: 8),
                        itemBuilder: (context, index) {
                          final user = users[index];
                          return _AdminUserRow(
                            user: user,
                            onRoleChanged: (role) => _updateRole(user, role),
                            onDisable: () => _disableUser(user),
                            onRevokeSessions: () => _revokeSessions(user),
                          );
                        },
                      );
                    },
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _InviteUserCard extends StatelessWidget {
  const _InviteUserCard({
    required this.emailController,
    required this.nameController,
    required this.role,
    required this.submitting,
    required this.onRoleChanged,
    required this.onSubmit,
  });

  final TextEditingController emailController;
  final TextEditingController nameController;
  final String role;
  final bool submitting;
  final ValueChanged<String?> onRoleChanged;
  final VoidCallback onSubmit;

  @override
  Widget build(BuildContext context) {
    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: LayoutBuilder(
          builder: (context, constraints) {
            final fields = [
              Expanded(
                flex: 2,
                child: TextField(
                  key: const Key('invite-email'),
                  controller: emailController,
                  decoration: const InputDecoration(
                    labelText: '邮箱',
                    isDense: true,
                  ),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                flex: 2,
                child: TextField(
                  key: const Key('invite-name'),
                  controller: nameController,
                  decoration: const InputDecoration(
                    labelText: '姓名',
                    isDense: true,
                  ),
                ),
              ),
              const SizedBox(width: 12),
              DropdownButton<String>(
                value: role,
                items: const [
                  DropdownMenuItem(value: 'viewer', child: Text('viewer')),
                  DropdownMenuItem(value: 'admin', child: Text('admin')),
                ],
                onChanged: submitting ? null : onRoleChanged,
              ),
              const SizedBox(width: 12),
              FilledButton.icon(
                onPressed: submitting ? null : onSubmit,
                icon: const Icon(Icons.person_add, size: 18),
                label: const Text('邀请'),
              ),
            ];

            if (constraints.maxWidth < 720) {
              return Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  TextField(
                    key: const Key('invite-email'),
                    controller: emailController,
                    decoration: const InputDecoration(labelText: '邮箱'),
                  ),
                  const SizedBox(height: 8),
                  TextField(
                    key: const Key('invite-name'),
                    controller: nameController,
                    decoration: const InputDecoration(labelText: '姓名'),
                  ),
                  const SizedBox(height: 8),
                  DropdownButton<String>(
                    value: role,
                    isExpanded: true,
                    items: const [
                      DropdownMenuItem(value: 'viewer', child: Text('viewer')),
                      DropdownMenuItem(value: 'admin', child: Text('admin')),
                    ],
                    onChanged: submitting ? null : onRoleChanged,
                  ),
                  const SizedBox(height: 8),
                  FilledButton.icon(
                    onPressed: submitting ? null : onSubmit,
                    icon: const Icon(Icons.person_add, size: 18),
                    label: const Text('邀请'),
                  ),
                ],
              );
            }

            return Row(children: fields);
          },
        ),
      ),
    );
  }
}

class _AdminUserRow extends StatelessWidget {
  const _AdminUserRow({
    required this.user,
    required this.onRoleChanged,
    required this.onDisable,
    required this.onRevokeSessions,
  });

  final ProviderUser user;
  final ValueChanged<String> onRoleChanged;
  final VoidCallback onDisable;
  final VoidCallback onRevokeSessions;

  @override
  Widget build(BuildContext context) {
    final disabled = user.status == 'disabled';
    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        child: Row(
          children: [
            Icon(
              disabled ? Icons.block : Icons.account_circle,
              color: disabled ? Colors.red.shade700 : const Color(0xFF2563EB),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    user.displayName,
                    style: const TextStyle(fontWeight: FontWeight.bold),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    '${user.email} · ${user.status}',
                    style: const TextStyle(
                      fontSize: 12,
                      color: Color(0xFF6B7280),
                    ),
                  ),
                ],
              ),
            ),
            DropdownButton<String>(
              value: user.role == 'admin' ? 'admin' : 'viewer',
              items: const [
                DropdownMenuItem(value: 'viewer', child: Text('viewer')),
                DropdownMenuItem(value: 'admin', child: Text('admin')),
              ],
              onChanged: disabled
                  ? null
                  : (role) {
                      if (role != null && role != user.role) {
                        onRoleChanged(role);
                      }
                    },
            ),
            const SizedBox(width: 8),
            IconButton(
              tooltip: '撤销会话',
              onPressed: onRevokeSessions,
              icon: const Icon(Icons.link_off),
            ),
            IconButton(
              tooltip: '停用账号',
              onPressed: disabled ? null : onDisable,
              icon: const Icon(Icons.person_off),
            ),
          ],
        ),
      ),
    );
  }
}

class BucketListView extends StatefulWidget {
  const BucketListView({
    super.key,
    required this.buckets,
    this.client,
    this.showMetadataOnlyBuckets = true,
    this.category,
    this.searchText = '',
    this.sortMode = BucketSortMode.none,
    this.sortAscending = true,
    this.createdAtDescending = true,
  });

  final Future<List<Bucket>>? buckets;
  final ProviderApiClient? client;
  final bool showMetadataOnlyBuckets;
  final String? category;
  final String searchText;
  final BucketSortMode sortMode;
  final bool sortAscending;
  final bool createdAtDescending;

  @override
  State<BucketListView> createState() => _BucketListViewState();
}

class _BucketListViewState extends State<BucketListView> {
  int _page = 0;

  @override
  void didUpdateWidget(covariant BucketListView oldWidget) {
    super.didUpdateWidget(oldWidget);
    // 分类、搜索、排序变化时回到第一页
    if (oldWidget.category != widget.category ||
        oldWidget.searchText != widget.searchText ||
        oldWidget.sortMode != widget.sortMode ||
        oldWidget.sortAscending != widget.sortAscending ||
        oldWidget.createdAtDescending != widget.createdAtDescending) {
      _page = 0;
    }
  }

  @override
  Widget build(BuildContext context) {
    return FutureBuilder<List<Bucket>>(
      future: widget.buckets,
      builder: (context, snapshot) {
        if (snapshot.connectionState != ConnectionState.done) {
          return const Center(child: CircularProgressIndicator());
        }
        if (snapshot.hasError) {
          return Center(
            child: Text(
              '加载桶失败：${snapshot.error}',
              style: TextStyle(color: Colors.red.shade700),
            ),
          );
        }
        var data = List<Bucket>.of(snapshot.data ?? const []);
        if (!widget.showMetadataOnlyBuckets) {
          data = visibleBucketsForRole('viewer', data);
        }

        // 分类过滤
        if (widget.category != null) {
          data = data.where((b) => b.category == widget.category).toList();
        }
        // 搜索过滤（按桶名）
        if (widget.searchText.isNotEmpty) {
          data = data
              .where((b) => b.name
                  .toLowerCase()
                  .contains(widget.searchText.toLowerCase()))
              .toList();
        }
        // 排序
        final sortByCreatedAt = widget.sortMode == BucketSortMode.createdAt ||
            widget.sortMode == BucketSortMode.createdAtThenName;
        final sortByName = widget.sortMode == BucketSortMode.name ||
            widget.sortMode == BucketSortMode.createdAtThenName;

        if (sortByCreatedAt || sortByName) {
          data.sort((a, b) {
            final nameCmp =
                a.name.toLowerCase().compareTo(b.name.toLowerCase());

            if (sortByCreatedAt) {
              final createdAtCmp = a.createdAt.compareTo(b.createdAt);
              if (createdAtCmp != 0) {
                return widget.createdAtDescending
                    ? -createdAtCmp
                    : createdAtCmp;
              }
              if (sortByName) {
                return widget.sortAscending ? nameCmp : -nameCmp;
              }
              return 0;
            }

            return widget.sortAscending ? nameCmp : -nameCmp;
          });
        }

        if (data.isEmpty) {
          return const Center(child: Text('没有匹配的桶'));
        }

        return LayoutBuilder(
          builder: (context, constraints) {
            final crossAxisCount = _bucketGridColumnCount(constraints.maxWidth);
            final pageSize = crossAxisCount * _bucketGridRowsPerPage;
            final totalPages = (data.length / pageSize).ceil();

            // 防止窗口变窄或筛选结果减少后页码越界。
            if (_page >= totalPages) {
              _page = totalPages - 1;
            }
            final start = _page * pageSize;
            final end = math.min(start + pageSize, data.length);
            final pageData = data.sublist(start, end);

            return Column(
              children: [
                Expanded(
                  child: GridView.builder(
                    itemCount: pageData.length,
                    gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
                      crossAxisCount: crossAxisCount,
                      mainAxisSpacing: _bucketGridSpacing,
                      crossAxisSpacing: _bucketGridSpacing,
                      mainAxisExtent: _bucketGridMainExtent,
                    ),
                    itemBuilder: (context, index) {
                      return BucketCard(
                        bucket: pageData[index],
                        client: widget.client,
                      );
                    },
                  ),
                ),
                const SizedBox(height: 8),
                _PaginationBar(
                  page: _page,
                  totalPages: totalPages,
                  onChanged: (newPage) {
                    setState(() {
                      _page = newPage;
                    });
                  },
                ),
              ],
            );
          },
        );
      },
    );
  }
}

class _PaginationBar extends StatelessWidget {
  const _PaginationBar({
    required this.page,
    required this.totalPages,
    required this.onChanged,
  });

  final int page;
  final int totalPages;
  final ValueChanged<int> onChanged;

  @override
  Widget build(BuildContext context) {
    if (totalPages <= 1) {
      return const SizedBox.shrink();
    }
    return Row(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        IconButton(
          tooltip: '上一页',
          onPressed: page > 0 ? () => onChanged(page - 1) : null,
          icon: const Icon(Icons.chevron_left, size: 20),
        ),
        Text(
          '${page + 1} / $totalPages',
          style: const TextStyle(fontSize: 13, color: Color(0xFF6B7280)),
        ),
        IconButton(
          tooltip: '下一页',
          onPressed: page < totalPages - 1 ? () => onChanged(page + 1) : null,
          icon: const Icon(Icons.chevron_right, size: 20),
        ),
      ],
    );
  }
}

class _CategoryFilterBar extends StatelessWidget {
  const _CategoryFilterBar({
    required this.categories,
    this.buckets,
    this.showMetadataOnlyBuckets = true,
    required this.selected,
    required this.onSelected,
  });

  final List<String> categories;
  final Future<List<Bucket>>? buckets;
  final bool showMetadataOnlyBuckets;
  final String? selected;
  final ValueChanged<String?> onSelected;

  @override
  Widget build(BuildContext context) {
    return FutureBuilder<List<Bucket>>(
      future: buckets,
      builder: (context, snapshot) {
        var data = List<Bucket>.of(snapshot.data ?? const []);
        if (!showMetadataOnlyBuckets) {
          data = visibleBucketsForRole('viewer', data);
        }
        return Wrap(
          spacing: 8,
          runSpacing: 8,
          children: [
            _FilterChip(
              label: '全部 ${data.length}',
              isSelected: selected == null,
              onTap: () => onSelected(null),
            ),
            for (final c in categories)
              _FilterChip(
                label: '$c ${data.where((b) => b.category == c).length}',
                isSelected: selected == c,
                onTap: () => onSelected(c),
              ),
          ],
        );
      },
    );
  }
}

class _FilterChip extends StatelessWidget {
  const _FilterChip({
    required this.label,
    required this.isSelected,
    required this.onTap,
  });

  final String label;
  final bool isSelected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(20),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        decoration: BoxDecoration(
          color: isSelected ? const Color(0xFF2563EB) : Colors.white,
          borderRadius: BorderRadius.circular(20),
          border: Border.all(
            color: isSelected ? const Color(0xFF2563EB) : Colors.grey.shade300,
          ),
        ),
        child: Text(
          label,
          style: TextStyle(
            fontSize: 13,
            fontWeight: FontWeight.w600,
            color: isSelected ? Colors.white : Colors.grey.shade700,
          ),
        ),
      ),
    );
  }
}

class BucketCard extends StatelessWidget {
  const BucketCard({super.key, required this.bucket, this.client});

  final Bucket bucket;
  final ProviderApiClient? client;

  @override
  Widget build(BuildContext context) {
    return Card(
      elevation: 0,
      clipBehavior: Clip.antiAlias,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
      child: InkWell(
        onTap: () {
          if (!bucket.canBrowseObjects) {
            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(content: Text('私密桶仅展示元数据')),
            );
            return;
          }
          Navigator.of(context).push(
            MaterialPageRoute(
              builder: (_) => BucketObjectsScreen(
                bucketName: bucket.name,
                client: client,
              ),
            ),
          );
        },
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 14),
          child: Row(
            children: [
              Container(
                width: 40,
                height: 40,
                decoration: BoxDecoration(
                  color: bucket.categoryColor.withValues(alpha: 0.12),
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Icon(bucket.categoryIcon,
                    color: bucket.categoryColor, size: 22),
              ),
              const SizedBox(width: 10),
              Expanded(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      bucket.name,
                      style: const TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.bold,
                      ),
                      overflow: TextOverflow.ellipsis,
                      maxLines: 1,
                    ),
                    const SizedBox(height: 4),
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 8,
                        vertical: 2,
                      ),
                      decoration: BoxDecoration(
                        color: bucket.categoryColor.withValues(alpha: 0.12),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Text(
                        bucket.category,
                        style: TextStyle(
                          fontSize: 11,
                          fontWeight: FontWeight.w600,
                          color: bucket.categoryColor,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _ProviderStatusCard extends StatelessWidget {
  const _ProviderStatusCard({required this.health});

  final Future<Map<String, dynamic>>? health;

  @override
  Widget build(BuildContext context) {
    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
        child: FutureBuilder<Map<String, dynamic>>(
          future: health,
          builder: (context, snapshot) {
            final hasError = snapshot.hasError;
            final data = snapshot.data;
            final status = hasError
                ? 'unavailable'
                : data?['status']?.toString() ?? 'checking';
            final service =
                data?['service']?.toString() ?? 'qtcloud-asset-provider';

            return Row(
              children: [
                Icon(
                  hasError ? Icons.error_outline : Icons.check_circle_outline,
                  size: 16,
                  color: hasError ? Colors.red.shade700 : Colors.green.shade700,
                ),
                const SizedBox(width: 6),
                Expanded(
                  child: Text(
                    '$service · $status',
                    style: TextStyle(
                      fontSize: 12,
                      color: Colors.grey.shade600,
                    ),
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
                if (hasError)
                  Tooltip(
                    message: snapshot.error.toString(),
                    child: Icon(
                      Icons.info_outline,
                      size: 14,
                      color: Colors.red.shade700,
                    ),
                  ),
              ],
            );
          },
        ),
      ),
    );
  }
}

class BucketObjectsScreen extends StatefulWidget {
  const BucketObjectsScreen({super.key, required this.bucketName, this.client});

  final String bucketName;
  final ProviderApiClient? client;

  @override
  State<BucketObjectsScreen> createState() => _BucketObjectsScreenState();
}

class _BucketObjectsScreenState extends State<BucketObjectsScreen> {
  late final ProviderApiClient _client;
  late Future<List<OssObject>> _objects;
  String _searchText = '';
  String? _sortKey; // 'date' | 'size' | null
  bool _dateDesc = true; // 日期默认新到旧
  bool _sizeDesc = true; // 大小默认大到小
  String _currentPrefix = ''; // 当前目录 prefix（'' = 根目录）

  @override
  void initState() {
    super.initState();
    _client = widget.client ?? ProviderApiClient();
    _objects = _fetchObjects();
  }

  /// 从全量对象中，解析出当前目录下的「直接子项」。
  /// 返回的每一项：文件保留原 key；目录用「第一层子目录路径」表示（以 / 结尾）。
  List<OssObject> _directChildren(List<OssObject> all) {
    final prefix = _currentPrefix;
    final files = <OssObject>[];
    final dirs = <String>{};

    for (final o in all) {
      final key = o.key;
      // 跳过不在当前目录下的对象
      if (!key.startsWith(prefix)) continue;
      // 去掉当前目录前缀，得到相对路径
      final rest = key.substring(prefix.length);
      if (rest.isEmpty) continue;

      final slash = rest.indexOf('/');
      if (slash == -1) {
        // 直接子文件
        files.add(o);
      } else {
        // 第一层子目录
        final dirName = rest.substring(0, slash + 1); // 含结尾 /
        dirs.add(prefix + dirName);
      }
    }

    final result = <OssObject>[];
    result.addAll(files);
    for (final d in dirs) {
      result.add(OssObject(
        key: d,
        size: 0,
        type: 'Directory',
        storageClass: '',
        lastModified: '',
      ));
    }
    return result;
  }

  /// 返回上一级目录 prefix。根目录的上级仍是根目录。
  String _parentPrefix(String prefix) {
    if (prefix.isEmpty) return '';
    final trimmed =
        prefix.endsWith('/') ? prefix.substring(0, prefix.length - 1) : prefix;
    final idx = trimmed.lastIndexOf('/');
    if (idx == -1) return '';
    return trimmed.substring(0, idx + 1);
  }

  /// 生成对象访问链接并复制到剪贴板。
  Future<void> _copyObjectUrl(OssObject object, int expiresSeconds) async {
    final link = await _client.fetchObjectUrl(
      bucketName: widget.bucketName,
      objectKey: object.key,
      expiresSeconds: expiresSeconds,
    );

    await Clipboard.setData(ClipboardData(text: link));
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('链接已复制：$link')),
      );
    }
  }

  /// 弹出有效期选择对话框，选完后生成链接。
  Future<void> _showExpiryDialog(OssObject object) async {
    final expiry = await showDialog<int>(
      context: context,
      builder: (context) {
        return SimpleDialog(
          title: const Text('选择链接有效期'),
          children: [
            _expiryOption(context, '1 天', 86400),
            _expiryOption(context, '7 天', 604800),
          ],
        );
      },
    );
    if (expiry == null) return; // 用户取消

    try {
      await _copyObjectUrl(object, expiry);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('生成链接失败：$e')),
        );
      }
    }
  }

  Widget _expiryOption(BuildContext context, String label, int seconds) {
    return SimpleDialogOption(
      onPressed: () => Navigator.of(context).pop(seconds),
      child: Text(label),
    );
  }

  Future<List<OssObject>> _fetchObjects() async {
    return _client.fetchObjects(widget.bucketName);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF7F8FA),
      appBar: AppBar(
        title: Text(_currentPrefix.isEmpty
            ? widget.bucketName
            : '${widget.bucketName}/$_currentPrefix'),
      ),
      body: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 960),
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                if (_currentPrefix.isNotEmpty) ...[
                  TextButton.icon(
                    onPressed: () {
                      setState(() {
                        _currentPrefix = _parentPrefix(_currentPrefix);
                      });
                    },
                    icon: const Icon(Icons.arrow_upward, size: 18),
                    label: const Text('返回上级'),
                    style: TextButton.styleFrom(
                      foregroundColor: const Color(0xFF2563EB),
                    ),
                  ),
                  const SizedBox(height: 4),
                ],
                Row(
                  children: [
                    Expanded(
                      child: TextField(
                        onChanged: (value) {
                          setState(() {
                            _searchText = value.trim();
                          });
                        },
                        decoration: InputDecoration(
                          hintText: '搜索文件…',
                          prefixIcon: const Icon(Icons.search, size: 20),
                          isDense: true,
                          contentPadding: const EdgeInsets.symmetric(
                            horizontal: 12,
                            vertical: 10,
                          ),
                          filled: true,
                          fillColor: Colors.white,
                          border: OutlineInputBorder(
                            borderRadius: BorderRadius.circular(8),
                            borderSide: BorderSide(color: Colors.grey.shade300),
                          ),
                          enabledBorder: OutlineInputBorder(
                            borderRadius: BorderRadius.circular(8),
                            borderSide: BorderSide(color: Colors.grey.shade300),
                          ),
                        ),
                      ),
                    ),
                    const SizedBox(width: 8),
                    _SortButton(
                      label: '日期',
                      desc: _dateDesc,
                      active: _sortKey == 'date',
                      onTap: () {
                        setState(() {
                          if (_sortKey == 'date') {
                            _dateDesc = !_dateDesc;
                          } else {
                            _sortKey = 'date';
                          }
                        });
                      },
                    ),
                    const SizedBox(width: 8),
                    _SortButton(
                      label: '大小',
                      desc: _sizeDesc,
                      active: _sortKey == 'size',
                      onTap: () {
                        setState(() {
                          if (_sortKey == 'size') {
                            _sizeDesc = !_sizeDesc;
                          } else {
                            _sortKey = 'size';
                          }
                        });
                      },
                    ),
                  ],
                ),
                const SizedBox(height: 16),
                Expanded(
                  child: FutureBuilder<List<OssObject>>(
                    future: _objects,
                    builder: (context, snapshot) {
                      if (snapshot.connectionState != ConnectionState.done) {
                        return const Center(child: CircularProgressIndicator());
                      }
                      if (snapshot.hasError) {
                        return Center(
                          child: Text(
                            '加载文件失败：${snapshot.error}',
                            style: TextStyle(color: Colors.red.shade700),
                          ),
                        );
                      }
                      // 先按当前目录过滤出「直接子项」
                      var data = _directChildren(snapshot.data ?? []);

                      // 搜索过滤
                      if (_searchText.isNotEmpty) {
                        data = data
                            .where((o) => o.key
                                .toLowerCase()
                                .contains(_searchText.toLowerCase()))
                            .toList();
                      }
                      // 排序
                      if (_sortKey == 'date') {
                        data.sort((a, b) {
                          final cmp = a.lastModified.compareTo(b.lastModified);
                          return _dateDesc ? -cmp : cmp;
                        });
                      } else if (_sortKey == 'size') {
                        data.sort((a, b) {
                          final cmp = a.size.compareTo(b.size);
                          return _sizeDesc ? -cmp : cmp;
                        });
                      }

                      if (data.isEmpty) {
                        return const Center(child: Text('没有匹配的文件'));
                      }
                      return ListView.builder(
                        itemCount: data.length,
                        itemBuilder: (context, index) {
                          final obj = data[index];
                          return _ObjectRow(
                            object: obj,
                            onTap: obj.isDir
                                ? () {
                                    setState(() {
                                      _currentPrefix = obj.key;
                                    });
                                  }
                                : null,
                            onCopyLink: obj.isDir ||
                                    !canExposeObjectLinks(widget.bucketName)
                                ? null
                                : () => _showExpiryDialog(obj),
                          );
                        },
                      );
                    },
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _SortButton extends StatelessWidget {
  const _SortButton({
    required this.label,
    required this.desc,
    required this.active,
    required this.onTap,
  });

  final String label;
  final bool desc;
  final bool active;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(8),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
        decoration: BoxDecoration(
          color: active
              ? const Color(0xFF2563EB).withValues(alpha: 0.1)
              : Colors.white,
          borderRadius: BorderRadius.circular(8),
          border: Border.all(
            color: active ? const Color(0xFF2563EB) : Colors.grey.shade300,
          ),
        ),
        child: Row(
          children: [
            Text(
              label,
              style: TextStyle(
                fontSize: 13,
                fontWeight: FontWeight.w600,
                color: active ? const Color(0xFF2563EB) : Colors.grey.shade700,
              ),
            ),
            const SizedBox(width: 4),
            Icon(
              desc ? Icons.arrow_downward : Icons.arrow_upward,
              size: 16,
              color: active ? const Color(0xFF2563EB) : Colors.grey.shade600,
            ),
          ],
        ),
      ),
    );
  }
}

class _ObjectRow extends StatelessWidget {
  const _ObjectRow({required this.object, this.onTap, this.onCopyLink});

  final OssObject object;
  final VoidCallback? onTap;
  final VoidCallback? onCopyLink;

  @override
  Widget build(BuildContext context) {
    final isDir = object.isDir;
    return Card(
      elevation: 0,
      clipBehavior: Clip.antiAlias,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
      child: InkWell(
        onTap: onTap,
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
          child: Row(
            children: [
              Icon(
                isDir ? Icons.folder : Icons.insert_drive_file,
                color: isDir ? const Color(0xFFF59E0B) : Colors.grey.shade600,
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Text(
                  // 目录显示去掉前缀、只显示目录名
                  isDir
                      ? '${object.key.split('/').where((s) => s.isNotEmpty).last}/'
                      : object.key,
                  style: const TextStyle(fontSize: 14),
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              if (!isDir)
                Text(
                  object.sizeLabel,
                  style: TextStyle(color: Colors.grey.shade500, fontSize: 12),
                ),
              const SizedBox(width: 16),
              Text(
                object.lastModified,
                style: TextStyle(color: Colors.grey.shade500, fontSize: 12),
              ),
              const SizedBox(width: 8),
              if (onCopyLink != null)
                IconButton(
                  tooltip: '复制链接',
                  onPressed: onCopyLink,
                  icon: const Icon(Icons.link, size: 18),
                  color: const Color(0xFF2563EB),
                  visualDensity: VisualDensity.compact,
                  padding: EdgeInsets.zero,
                  constraints: const BoxConstraints(),
                ),
              if (isDir) const Icon(Icons.chevron_right, color: Colors.grey),
            ],
          ),
        ),
      ),
    );
  }
}
