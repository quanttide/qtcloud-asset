import 'dart:typed_data';
import 'dart:js_interop';

import 'package:web/web.dart' as web;

Future<void> triggerDownloadImpl(Uri url, String fileName) async {
  final body = web.document.body;
  if (body == null) {
    throw StateError('browser document is not ready');
  }

  final anchor = web.HTMLAnchorElement()
    ..href = url.toString()
    ..download = fileName
    ..style.display = 'none';
  body.append(anchor);
  try {
    anchor.click();
  } finally {
    anchor.remove();
  }
}

Future<void> triggerBytesDownloadImpl(
  List<int> bytes,
  String fileName,
) async {
  final data = Uint8List.fromList(bytes).toJS;
  final blob = web.Blob(
    <web.BlobPart>[data].toJS,
    web.BlobPropertyBag(type: 'application/zip'),
  );
  final url = web.URL.createObjectURL(blob);
  try {
    await triggerDownloadImpl(Uri.parse(url), fileName);
  } finally {
    web.URL.revokeObjectURL(url);
  }
}
