#!/bin/sh
# Создаёт топики из /contracts/topics.yaml (смонтирован из contracts/topics.yaml).
# Ожидаемый формат:
#   topics:
#     - name: posts.events
#       partitions: 12
#       replication: 1
# LF-переводы строк обязательны.
set -eu

BOOTSTRAP="${KAFKA_BOOTSTRAP:-kafka:9092}"
FILE="${TOPICS_FILE:-/contracts/topics.yaml}"

awk '
  /^[[:space:]]*- name:/ { if (name != "") print name, p, r; name = $3; p = 1; r = 1 }
  /^[[:space:]]*partitions:/  { p = $2 }
  /^[[:space:]]*replication:/ { r = $2 }
  END { if (name != "") print name, p, r }
' "$FILE" | while read -r name partitions replication; do
  echo "topic ${name}: partitions=${partitions} replication=${replication}"
  /opt/kafka/bin/kafka-topics.sh --bootstrap-server "$BOOTSTRAP" \
    --create --if-not-exists \
    --topic "$name" --partitions "$partitions" --replication-factor "$replication"
done
