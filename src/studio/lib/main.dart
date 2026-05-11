import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;

const providerBaseUrl = 'https://api.asset.quanttide.com';

void main() {
  runApp(const QtCloudAssetStudio());
}

class QtCloudAssetStudio extends StatelessWidget {
  const QtCloudAssetStudio({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'QtCloud Asset',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: const Color(0xFF2563EB)),
        useMaterial3: true,
      ),
      home: const DashboardScreen(),
    );
  }
}

class DashboardScreen extends StatefulWidget {
  const DashboardScreen({super.key});

  @override
  State<DashboardScreen> createState() => _DashboardScreenState();
}

class _DashboardScreenState extends State<DashboardScreen> {
  Future<Map<String, dynamic>>? _health;

  @override
  void initState() {
    super.initState();
    _health = _fetchHealth();
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

  void _refresh() {
    setState(() {
      _health = _fetchHealth();
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF7F8FA),
      appBar: AppBar(
        title: const Text('QtCloud Asset'),
        actions: [
          IconButton(
            tooltip: 'Refresh provider status',
            onPressed: _refresh,
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
                Text(
                  'Digital Asset Console',
                  style: Theme.of(context).textTheme.headlineMedium,
                ),
                const SizedBox(height: 8),
                const Text('Studio is hosted on OSS and connected to the FC provider.'),
                const SizedBox(height: 24),
                _ProviderStatusCard(health: _health),
              ],
            ),
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
        padding: const EdgeInsets.all(20),
        child: FutureBuilder<Map<String, dynamic>>(
          future: health,
          builder: (context, snapshot) {
            final hasError = snapshot.hasError;
            final data = snapshot.data;
            final status = hasError ? 'unavailable' : data?['status']?.toString() ?? 'checking';
            final service = data?['service']?.toString() ?? 'qtcloud-asset-provider';

            return Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Icon(
                      hasError ? Icons.error_outline : Icons.check_circle_outline,
                      color: hasError ? Colors.red.shade700 : Colors.green.shade700,
                    ),
                    const SizedBox(width: 8),
                    Text(
                      'Provider status',
                      style: Theme.of(context).textTheme.titleLarge,
                    ),
                  ],
                ),
                const SizedBox(height: 16),
                _StatusRow(label: 'Service', value: service),
                _StatusRow(label: 'Status', value: status),
                _StatusRow(label: 'Endpoint', value: providerBaseUrl),
                if (hasError) ...[
                  const SizedBox(height: 12),
                  Text(
                    snapshot.error.toString(),
                    style: TextStyle(color: Colors.red.shade700),
                  ),
                ],
              ],
            );
          },
        ),
      ),
    );
  }
}

class _StatusRow extends StatelessWidget {
  const _StatusRow({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 6),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 92,
            child: Text(
              label,
              style: const TextStyle(fontWeight: FontWeight.w600),
            ),
          ),
          Expanded(child: Text(value)),
        ],
      ),
    );
  }
}
