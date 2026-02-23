ARG SEARCH_VERSION=7.17.10
FROM docker.elastic.co/elasticsearch/elasticsearch:${SEARCH_VERSION}

RUN bin/elasticsearch-plugin install --batch analysis-icu || true
RUN bin/elasticsearch-plugin install --batch analysis-phonetic || true
