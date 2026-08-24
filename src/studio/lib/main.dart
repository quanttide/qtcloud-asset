import 'dart:convert';
import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:http/http.dart' as http;

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

bool canExposeObjectLinks(String bucketName) {
  return !bucketName.endsWith('-private') &&
      bucketName != 'quanttide-terraform-state';
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
  const QtCloudAssetStudio({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: '量潮资产云',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: const Color(0xFF2563EB)),
        useMaterial3: true,
      ),
      home: const DashboardScreen(),
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
  const DashboardScreen({super.key});

  @override
  State<DashboardScreen> createState() => _DashboardScreenState();
}

class _DashboardScreenState extends State<DashboardScreen> {
  Future<Map<String, dynamic>>? _health;
  Future<List<Bucket>>? _buckets;
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
    _health = _fetchHealth();
    _buckets = _fetchBuckets();
  }

  Future<Map<String, dynamic>> _fetchHealth() async {
    final response = await http
        .get(Uri.parse('$providerBaseUrl/health'))
        .timeout(const Duration(seconds: 15));
    if (response.statusCode != 200) {
      throw Exception('Provider returned HTTP ${response.statusCode}');
    }
    return jsonDecode(response.body) as Map<String, dynamic>;
  }

  Future<List<Bucket>> _fetchBuckets() async {
    final response = await http
        .get(Uri.parse('$providerBaseUrl/buckets'))
        .timeout(const Duration(seconds: 15));
    if (response.statusCode != 200) {
      throw Exception('Provider returned HTTP ${response.statusCode}');
    }
    final body = jsonDecode(response.body) as Map<String, dynamic>;
    final list = body['buckets'] as List<dynamic>;
    return list.map((e) => Bucket.fromJson(e as Map<String, dynamic>)).toList();
  }

  void _refresh() {
    setState(() {
      _health = _fetchHealth();
      _buckets = _fetchBuckets();
    });
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
    final createdAtLabel =
        _createdAtSortEnabled ? '创建时间' : '创建时间：关闭';
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
    return Scaffold(
      backgroundColor: const Color(0xFFF7F8FA),
      appBar: AppBar(
        title: const Text('量潮资产云'),
        actions: [
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
            padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 24),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  '对象存储桶',
                  style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
                ),
                const SizedBox(height: 12),
                _CategoryFilterBar(
                  categories: _categories,
                  buckets: _buckets,
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
  }
}

class BucketListView extends StatefulWidget {
  const BucketListView({
    super.key,
    required this.buckets,
    this.category,
    this.searchText = '',
    this.sortMode = BucketSortMode.none,
    this.sortAscending = true,
    this.createdAtDescending = true,
  });

  final Future<List<Bucket>>? buckets;
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
                      return BucketCard(bucket: pageData[index]);
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
    required this.selected,
    required this.onSelected,
  });

  final List<String> categories;
  final Future<List<Bucket>>? buckets;
  final String? selected;
  final ValueChanged<String?> onSelected;

  @override
  Widget build(BuildContext context) {
    return FutureBuilder<List<Bucket>>(
      future: buckets,
      builder: (context, snapshot) {
        final data = snapshot.data ?? [];
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
  const BucketCard({super.key, required this.bucket});

  final Bucket bucket;

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
              builder: (_) => BucketObjectsScreen(bucketName: bucket.name),
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
  const BucketObjectsScreen({super.key, required this.bucketName});

  final String bucketName;

  @override
  State<BucketObjectsScreen> createState() => _BucketObjectsScreenState();
}

class _BucketObjectsScreenState extends State<BucketObjectsScreen> {
  late Future<List<OssObject>> _objects;
  String _searchText = '';
  String? _sortKey; // 'date' | 'size' | null
  bool _dateDesc = true; // 日期默认新到旧
  bool _sizeDesc = true; // 大小默认大到小
  String _currentPrefix = ''; // 当前目录 prefix（'' = 根目录）

  @override
  void initState() {
    super.initState();
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
    final bucket = Uri.encodeComponent(widget.bucketName);
    final url =
        Uri.parse('$providerBaseUrl/buckets/$bucket/object-url').replace(
      queryParameters: {
        'key': object.key,
        'expires': '$expiresSeconds',
      },
    );

    final response = await http.get(url).timeout(const Duration(seconds: 20));
    if (response.statusCode != 200) {
      throw Exception('Provider returned HTTP ${response.statusCode}');
    }
    final body = jsonDecode(response.body) as Map<String, dynamic>;
    final link = body['url'] as String;

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
            _expiryOption(context, '30 天', 2592000),
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
    final name = Uri.encodeComponent(widget.bucketName);
    return fetchAllObjectPages((marker) async {
      final uri = Uri.parse('$providerBaseUrl/buckets/$name/objects').replace(
        queryParameters: marker.isEmpty ? {} : {'marker': marker},
      );
      final response = await http.get(uri).timeout(
            const Duration(seconds: 20),
          );
      if (response.statusCode != 200) {
        throw Exception('Provider returned HTTP ${response.statusCode}');
      }
      return jsonDecode(response.body) as Map<String, dynamic>;
    });
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
