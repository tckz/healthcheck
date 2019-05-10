out=results.bin

date
echo "GET https://example.com/" | \
    vegeta attack -rate=100/s -duration 3m > $out && \
    vegeta report < $out

