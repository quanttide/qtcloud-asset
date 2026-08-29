import 'dart:convert';
import 'dart:typed_data';

class StoredZipEntry {
  const StoredZipEntry({
    required this.name,
    required this.bytes,
  });

  final String name;
  final List<int> bytes;
}

List<int> buildStoredZip(Iterable<StoredZipEntry> entries) {
  final local = BytesBuilder(copy: false);
  final central = BytesBuilder(copy: false);
  final names = <String>{};
  var offset = 0;
  var count = 0;

  for (final entry in entries) {
    final name = _normalizeZipEntryName(entry.name);
    if (!names.add(name)) {
      throw ArgumentError('duplicate ZIP entry: $name');
    }
    final nameBytes = utf8.encode(name);
    if (nameBytes.length > 0xffff) {
      throw ArgumentError('ZIP entry name is too long');
    }
    if (entry.bytes.length > 0xffffffff) {
      throw ArgumentError('ZIP entry is too large');
    }

    final size = entry.bytes.length;
    final crc = _crc32(entry.bytes);
    final localHeader = BytesBuilder(copy: false);
    _u32(localHeader, 0x04034b50);
    _u16(localHeader, 20);
    _u16(localHeader, 0x0800);
    _u16(localHeader, 0);
    _u16(localHeader, 0);
    _u16(localHeader, 0);
    _u32(localHeader, crc);
    _u32(localHeader, size);
    _u32(localHeader, size);
    _u16(localHeader, nameBytes.length);
    _u16(localHeader, 0);
    local.add(localHeader.takeBytes());
    local.add(nameBytes);
    local.add(entry.bytes);

    final centralHeader = BytesBuilder(copy: false);
    _u32(centralHeader, 0x02014b50);
    _u16(centralHeader, 20);
    _u16(centralHeader, 20);
    _u16(centralHeader, 0x0800);
    _u16(centralHeader, 0);
    _u16(centralHeader, 0);
    _u16(centralHeader, 0);
    _u32(centralHeader, crc);
    _u32(centralHeader, size);
    _u32(centralHeader, size);
    _u16(centralHeader, nameBytes.length);
    _u16(centralHeader, 0);
    _u16(centralHeader, 0);
    _u16(centralHeader, 0);
    _u16(centralHeader, 0);
    _u32(centralHeader, 0);
    _u32(centralHeader, offset);
    central.add(centralHeader.takeBytes());
    central.add(nameBytes);

    offset += 30 + nameBytes.length + size;
    count++;
  }

  if (count > 0xffff) {
    throw ArgumentError('ZIP contains too many entries');
  }
  if (offset > 0xffffffff || central.length > 0xffffffff) {
    throw ArgumentError('ZIP is too large');
  }

  final centralBytes = central.takeBytes();
  final result = BytesBuilder(copy: false);
  result.add(local.takeBytes());
  result.add(centralBytes);
  _u32(result, 0x06054b50);
  _u16(result, 0);
  _u16(result, 0);
  _u16(result, count);
  _u16(result, count);
  _u32(result, centralBytes.length);
  _u32(result, offset);
  _u16(result, 0);
  return result.takeBytes();
}

String _normalizeZipEntryName(String name) {
  final parts = <String>[];
  for (final part in name.replaceAll(r'\', '/').split('/')) {
    if (part.isEmpty || part == '.') continue;
    if (part == '..') {
      if (parts.isNotEmpty) parts.removeLast();
      continue;
    }
    parts.add(part);
  }
  if (parts.isEmpty) {
    throw ArgumentError('ZIP entry name is empty');
  }
  return parts.join('/');
}

int _crc32(List<int> bytes) {
  var crc = 0xffffffff;
  for (final byte in bytes) {
    crc = _crcTable[(crc ^ byte) & 0xff] ^ (crc >> 8);
  }
  return (crc ^ 0xffffffff) & 0xffffffff;
}

void _u16(BytesBuilder output, int value) {
  output.add(<int>[
    value & 0xff,
    (value >> 8) & 0xff,
  ]);
}

void _u32(BytesBuilder output, int value) {
  output.add(<int>[
    value & 0xff,
    (value >> 8) & 0xff,
    (value >> 16) & 0xff,
    (value >> 24) & 0xff,
  ]);
}

final List<int> _crcTable = List<int>.generate(256, (index) {
  var value = index;
  for (var bit = 0; bit < 8; bit++) {
    value = (value & 1) == 1 ? 0xedb88320 ^ (value >> 1) : value >> 1;
  }
  return value;
});
