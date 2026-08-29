import 'download_url_io.dart'
    if (dart.library.js_interop) 'download_url_web.dart';

typedef DownloadUrlHandler = Future<void> Function(
  Uri url,
  String fileName,
);

typedef DownloadBytesHandler = Future<void> Function(
  List<int> bytes,
  String fileName,
);

Future<void> triggerDownload(Uri url, String fileName) {
  return triggerDownloadImpl(url, fileName);
}

Future<void> triggerBytesDownload(List<int> bytes, String fileName) {
  return triggerBytesDownloadImpl(bytes, fileName);
}
