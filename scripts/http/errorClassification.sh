#!/bin/bash
#Output filename: error_classification
#create tmp file
tmpName=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1).tmp
grep -v '"Error":""' $1 > $tmpName
total=$(wc -l < $tmpName)
unexpectedEOF=$(grep 'unexpected EOF' $tmpName | wc -l)
tooManyColons=$(grep 'dial tcp: too many colons in address' $tmpName | wc -l)
resetByPeer=$(grep 'connection reset by peer' $tmpName | wc -l)
stoppedAfterRedirects=$(grep 'stopped after 5 redirects' $tmpName | wc -l)
oversizedTLSRecord=$(grep 'tls: oversized record received with length' $tmpName | wc -l)
malformedHTTPResponse=$(grep 'malformed HTTP response' $tmpName | wc -l)
noSuchHost=$(grep 'no such host' $tmpName | wc -l)
ioTimeout=$(grep 'i/o timeout' $tmpName | wc -l)
malformedHTTPVersion=$(grep 'malformed HTTP version ' $tmpName | wc -l)
malformedHTTPStatusCode=$(grep 'malformed HTTP status code' $tmpName | wc -l)
getsocketpConnectionRefused=$(grep 'getsockopt: connection refused' $tmpName | wc -l)
eof=$(grep ': EOF' $tmpName | wc -l)
malformedMime=$(grep 'malformed MIME header line' $tmpName | wc -l)
strconvParsing=$(grep 'strconv.ParseUint: parsing' $tmpName | wc -l)
unsupportedEncoding=$(grep 'unsupported transfer encoding' $tmpName | wc -l)
gzipInvalidHeader=$(grep 'gzip: invalid header' $tmpName | wc -l)
remoteInternalError=$(grep 'remote error: internal error' $tmpName | wc -l)
handshakeError=$(grep 'remote error: handshake' $tmpName | wc -l)

extracted=$((handshakeError+gzipInvalidHeader+remoteInternalError+unsupportedEncoding+strconvParsing+malformedMime+eof+unexpectedEOF+tooManyColons+resetByPeer+stoppedAfterRedirects+oversizedTLSRecord+malformedHTTPResponse+noSuchHost+ioTimeout+malformedHTTPVersion+malformedHTTPStatusCode+getsocketpConnectionRefused))

rm $tmpName
printf '%s\n' "-----------Total errors: $total------------"
printf '%s\n' "\"unexpected EOF\": $unexpectedEOF ($(awk "BEGIN {printf \"%.2f\n\", 100/$total*$unexpectedEOF}"))%"
printf '%s\n' "\"too many colons in address\": $tooManyColons ($(awk "BEGIN {printf \"%.2f\n\", 100/$total*$tooManyColons}"))%"
printf '%s\n' "\"connection reset by peer\": $resetByPeer ($(awk "BEGIN {printf \"%.2f\n\", 100/$total*$resetByPeer}"))%"
printf '%s\n' "\"stopped after 5 redirects\": $stoppedAfterRedirects ($(awk "BEGIN {printf \"%.2f\n\", 100/$total*$stoppedAfterRedirects}"))%"
printf '%s\n' "\"tls: oversized record received with length\": $oversizedTLSRecord ($(awk "BEGIN {printf \"%.2f\n\", 100/$total*$oversizedTLSRecord}"))%"
printf '%s\n' "\"malformed HTTP response\": $malformedHTTPResponse ($(awk "BEGIN {printf \"%.2f\n\", 100/$total*$malformedHTTPResponse}"))%"
printf '%s\n' "\"no such host\": $noSuchHost ($(awk "BEGIN {printf \"%.2f\n\", 100/$total*$noSuchHost}"))%"
printf '%s\n' "\"i/o timeout\": $ioTimeout ($(awk "BEGIN {printf \"%.2f\n\", 100/$total*$ioTimeout}"))%"
printf '%s\n' "\"malformed HTTP version\": $malformedHTTPVersion ($(awk "BEGIN {printf \"%.2f\n\", 100/$total*$malformedHTTPVersion}"))%"
printf '%s\n' "\"malformed HTTP status code\": $malformedHTTPStatusCode ($(awk "BEGIN {printf \"%.2f\n\", 100/$total*$malformedHTTPStatusCode}"))%"
printf '%s\n' "\"getsockopt: connection refused\": $getsocketpConnectionRefused ($(awk "BEGIN {printf \"%.2f\n\", 100/$total*$getsocketpConnectionRefused}"))%"
printf '%s\n' "\"EOF\": $eof ($(awk "BEGIN {printf \"%.2f\n\", 100/$total*$eof}"))%"
printf '%s\n' "\"malformed MIME header line\": $malformedMime ($(awk "BEGIN {printf \"%.2f\n\", 100/$total*$malformedMime}"))%"
printf '%s\n' "\"strconv.ParseUint: parsing ... invalid syntax\": $strconvParsing ($(awk "BEGIN {printf \"%.2f\n\", 100/$total*$strconvParsing}"))%"
printf '%s\n' "\"unsupported transfer encoding\": $unsupportedEncoding ($(awk "BEGIN {printf \"%.2f\n\", 100/$total*$unsupportedEncoding}"))%"
printf '%s\n' "\"gzip: invalid header\": $gzipInvalidHeader ($(awk "BEGIN {printf \"%.2f\n\", 100/$total*$gzipInvalidHeader}"))%"
printf '%s\n' "\"remote error: internal error\": $remoteInternalError ($(awk "BEGIN {printf \"%.2f\n\", 100/$total*$remoteInternalError}"))%"
printf '%s\n' "\"TLS handshake error\": $handshakeError ($(awk "BEGIN {printf \"%.2f\n\", 100/$total*$handshakeError}"))%"

printf '%s\n' "-----------Extracted errors: $extracted ($(awk "BEGIN {printf \"%.2f\n\", 100/$total*$extracted}"))%------------"
printf '%s\n' "-----------Not extracted errors: $((total-extracted)) ($(awk "BEGIN {printf \"%.2f\n\", 100/$total*$((total-extracted))}"))%------------"

