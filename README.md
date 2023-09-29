# Terraform Provider for Apache Kafka

Kafka can be best managed by the original Kafka binaries. This provider is an attempt to reduce the effort of creating Kafka brokers, and does not have any broker management features.

## Limitations

- The provider is currently built for Debian-based Linux distributions only.
- Due to the way Kafka is built, the provider can only run 1 resource (cluster) at a time.
- The provider uses Kafka v3.5.0.
