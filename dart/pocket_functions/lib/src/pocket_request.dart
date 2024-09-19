import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';

import 'pocket_response.dart';

class PocketRequest {
  final String path;
  final String httpMethod;
  final Uint8List? body;
  final String contentType;
  final Map<String, List<String>> params;

  final response = PocketResponse();

  PocketRequest({
    required this.path,
    required this.httpMethod,
    required this.body,
    required this.contentType,
    required this.params,
  });

  factory PocketRequest.fromFile(String path) {
    var file = File(path);
    List<String> env  = [];

    final lineBuffer = <int>[];
    final readBody = false;

    var bytes = file.readAsBytesSync();
    for( var i=0; i< bytes.length; i++) {
      var byte = bytes[i];
      if (byte == 10 && !readBody) {
        String line = utf8.decode(lineBuffer);
        env.add(line);
        lineBuffer.clear();
      } else {
        lineBuffer.add(byte);
      }
    }

    var body = Uint8List.fromList(lineBuffer);
    Uri uri = Uri(query: env[1]);

    return PocketRequest(
      path: env[0],
      httpMethod: env[2],
      body: body,
      contentType: env[3],
      params: uri.queryParametersAll,
    );
  }

  String? plainBody() {
    return body.toString();
  }

  Map<String, dynamic>? jsonBody() {
    if (body == null) {
      return null;
    }

    if (contentType != "application/json") {
      return null;
    } else {
      return jsonDecode(String.fromCharCodes(body!));
    }
  }
}
