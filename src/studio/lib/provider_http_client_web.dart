import 'package:http/browser_client.dart';
import 'package:http/http.dart' as http;

http.Client createProviderHttpClientImpl() {
  return BrowserClient()..withCredentials = true;
}

http.Client createPublicObjectHttpClientImpl() => BrowserClient();
