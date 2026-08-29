Future<void> triggerDownloadImpl(Uri url, String fileName) async {
  throw UnsupportedError('file downloads are only supported on Flutter Web');
}

Future<void> triggerBytesDownloadImpl(
  List<int> bytes,
  String fileName,
) async {
  throw UnsupportedError('file downloads are only supported on Flutter Web');
}
