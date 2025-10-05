Cliente
  |
  | 1) POST /media/presign (namespace/scope/profile/entity → key)
  v
Publisher (API)
  |-- Inserta en Mongo:
  |      { key, status:"pending", urls, planned[], mediaType, ... }
  |-- Publica NATS (JetStream, stream=media):
  |      subject: media.pending.<key>
  v
═══════════════════════ NATS / JetStream (stream: "media") ═══════════════════════

          [Queue group: media-publisher]           [Queue group: media-transform]
          (varias réplicas de Publisher)           (varias réplicas de Consumer)
                |                                           |
2) Consume media.pending.<key>                              |
   (solo 1 réplica lo recibe, gracias a queue group)        |
                |                                           |
   HEAD S3 original                                         |
   ├─ OK → transición atómica en Mongo:                     |
   │       filter: {key, status:"pending"}                  |
   │       update: {$set:{status:"uploaded", mime, bytes, etag,...},
   │                $unset:{nextCheckAt, attempts}}         |
   │    └─ Publica: media.uploaded.<key>  ──────────────────┼────► 3) Consumer recibe media.uploaded.<key>
   └─ KO → reintentos con backoff + jitter                  |
           (y posible transición a failed)                  |
                                                            |
                                             3.1) LOCK/claim de trabajo:
                                                  transición atómica:
                                                  filter: {key, status:"uploaded"}
                                                  update: {$set:{status:"processing",
                                                                processingBy:<podId>,
                                                                startedAt: now}}
                                                  ├─ ModifiedCount==0 → otro pod ya lo tomó (ACK y salir)
                                                  └─ OK → sigue

                                             3.2) GET S3 original (con timeout)
                                             3.3) Ejecuta pipeline por mediaType usando planned[]:
                                                  - IMAGE: genera variantes WebP en tamaños del plan
                                                  - GIF: poster.webp y mp4_480 (si soportas)
                                                  - VIDEO: poster + mp4_480 + mp4_720 (si soportas)
                                             3.4) PUT S3 de variantes (+ Cache-Control, ACL)
                                             3.5) UPDATE Mongo:
                                                  {$set:{
                                                     variants[], checksum, width, height,
                                                     status:"ready", updatedAt: now,
                                                     pipelineVersion
                                                   }}
                                                  (o status:"failed" si algo salió mal)
                                             3.6) Publica:
                                                  - media.process.started.<key>   (opcional, al entrar a processing)
                                                  - media.process.completed.<key> (ready)
                                                    ó media.process.failed.<key>  (failed)

                                                 (Opcional: notify.* para UI en tiempo real)





