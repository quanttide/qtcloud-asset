import 'package:http/http.dart' as http;

import 'provider_http_client_io.dart'
    if (dart.library.js_interop) 'provider_http_client_web.dart';

http.Client createProviderHttpClient() => createProviderHttpClientImpl();

http.Client createPublicObjectHttpClient() =>
    createPublicObjectHttpClientImpl();
