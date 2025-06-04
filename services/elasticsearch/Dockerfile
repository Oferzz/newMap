# Use official Elasticsearch image
FROM elasticsearch:8.11.0

# Set environment variables for single-node deployment
ENV discovery.type=single-node
ENV xpack.security.enabled=false
ENV xpack.security.http.ssl.enabled=false
ENV xpack.security.transport.ssl.enabled=false
ENV ES_JAVA_OPTS="-Xms512m -Xmx512m"
ENV cluster.name=newmap-cluster
ENV node.name=newmap-node

# Create elasticsearch user if it doesn't exist
USER root
RUN groupadd -g 1000 elasticsearch 2>/dev/null || true && \
    useradd -u 1000 -g elasticsearch -s /bin/bash -m elasticsearch 2>/dev/null || true

# Create necessary directories with proper permissions
RUN mkdir -p /usr/share/elasticsearch/data && \
    chown -R elasticsearch:elasticsearch /usr/share/elasticsearch

# Switch to elasticsearch user
USER elasticsearch

# Expose Elasticsearch ports
EXPOSE 9200 9300

# Add healthcheck
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
  CMD curl -f http://localhost:9200/_cluster/health || exit 1

# Use the default entrypoint
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
CMD ["elasticsearch"]