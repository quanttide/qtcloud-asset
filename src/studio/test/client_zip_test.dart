import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:qtcloud_studio/client_zip.dart';

void main() {
  test('buildStoredZip creates a readable stored archive', () {
    final archive = buildStoredZip([
      StoredZipEntry(name: 'docs/readme.md', bytes: utf8.encode('hello')),
      StoredZipEntry(name: 'docs/guide.txt', bytes: utf8.encode('guide')),
    ]);

    final entries = _readStoredZip(archive);
    expect(entries, {
      'docs/readme.md': 'hello',
      'docs/guide.txt': 'guide',
    });
  });

  test('buildStoredZip rejects unsafe entry paths', () {
    expect(
      () => buildStoredZip([
        const StoredZipEntry(
          name: r'../../docs/./readme.md',
          bytes: [1, 2, 3],
        ),
      ]),
      throwsArgumentError,
    );
    expect(
      () => buildStoredZip([
        const StoredZipEntry(name: r'docs\readme.md', bytes: [1, 2, 3]),
      ]),
      throwsArgumentError,
    );
  });
}

Map<String, String> _readStoredZip(List<int> archive) {
  final bytes = archive;
  final entries = <String, String>{};
  var offset = 0;
  while (_uint32(bytes, offset) == 0x04034b50) {
    final nameLength = _uint16(bytes, offset + 26);
    final extraLength = _uint16(bytes, offset + 28);
    final size = _uint32(bytes, offset + 18);
    final nameStart = offset + 30;
    final dataStart = nameStart + nameLength + extraLength;
    final name = utf8.decode(bytes.sublist(nameStart, nameStart + nameLength));
    final data = bytes.sublist(dataStart, dataStart + size);
    entries[name] = utf8.decode(data, allowMalformed: true);
    offset = dataStart + size;
  }

  final centralOffset = offset;
  var centralCount = 0;
  while (_uint32(bytes, offset) == 0x02014b50) {
    final nameLength = _uint16(bytes, offset + 28);
    final extraLength = _uint16(bytes, offset + 30);
    final commentLength = _uint16(bytes, offset + 32);
    final localOffset = _uint32(bytes, offset + 42);
    expect(_uint32(bytes, localOffset), 0x04034b50);
    offset += 46 + nameLength + extraLength + commentLength;
    centralCount++;
  }

  expect(_uint32(bytes, offset), 0x06054b50);
  expect(_uint16(bytes, offset + 8), centralCount);
  expect(_uint16(bytes, offset + 10), centralCount);
  expect(_uint32(bytes, offset + 12), offset - centralOffset);
  expect(_uint32(bytes, offset + 16), centralOffset);
  return entries;
}

int _uint16(List<int> bytes, int offset) {
  return bytes[offset] | (bytes[offset + 1] << 8);
}

int _uint32(List<int> bytes, int offset) {
  return bytes[offset] |
      (bytes[offset + 1] << 8) |
      (bytes[offset + 2] << 16) |
      (bytes[offset + 3] << 24);
}
