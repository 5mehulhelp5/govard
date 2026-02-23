ARG SEARCH_VERSION=2.11.0
FROM opensearchproject/opensearch:${SEARCH_VERSION}

RUN bin/opensearch-plugin install --batch analysis-icu || true
RUN bin/opensearch-plugin install --batch analysis-phonetic || true
