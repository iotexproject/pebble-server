# record pebble device migration

| date       | timestamp(db updated_at)      | imei(s)                    | comment                                           |
|:-----------|:------------------------------|:---------------------------|:--------------------------------------------------|
| 2024-07-06 |                               | 103381234567407(simulator) | internal pebble                                   |
|            |                               | 351358810263647            |                                                   |
|            |                               | 351358813375182            |                                                   |
|            |                               | 351358810263514            |                                                   |
| 2024-07-11 | 2024-07-11 08:30:00(migrated) | 350916067079072            | the most active top 5 devices from 07-03 to 07-10 |
|            |                               | 350916067070162            |                                                   |
|            |                               | 350916067099906            |                                                   |
|            |                               | 350916067070345            |                                                   |
|            |                               | 351358810283462            |                                                   |

migrate steps:

1. prepare wasm with blacklist(devices in '2024-07-11')
2. publish wasm and restart old w3bstream
3. modify new pebble sequencer white list and restart new pebble sequencer(tag v0.2.3)

rollback steps:

1. stop new pebble sequencer
2. stop w3bstream
3. rollback wasm blacklist
4. export migrated device data from new pebble sequencer db
5. import migrated device data to old w3bstream wasm db
6. start w3bstream testnet
7. frontend switch to old db.

## export from 35.223.43.49(pebble-sequencer db)

```shell
PGPASSWORD=-------- \
psql -h 35.223.43.49 -U pebble -d pebble -c \
"\COPY (SELECT * FROM device where updated_at > '2024-07-11 08:30:00' and id in('350916067079072','350916067070162','350916067099906','350916067070345','351358810283462')) TO 'device.csv' CSV HEADER"

PGPASSWORD=-------- \
psql -h 35.223.43.49 -U pebble -d pebble -c \
"\COPY (SELECT * FROM device_record where updated_at > '2024-07-11 08:30:00' and imei in('350916067079072','350916067070162','350916067099906','350916067070345','351358810283462')) TO 'device_record.csv' CSV HEADER"
```

## import to 34.172.94.245(w3bstream pebble wasm db)

```shell
PGPASSWORD=-------- \
psql -h 34.172.94.245 -U w3bstream -d w3b_1456942923637714945 -c "\COPY device FROM 'device.csv' CSV HEADER"

PGPASSWORD=-------- \
psql -h 34.172.94.245 -U w3bstream -d w3b_1456942923637714945 -c "\COPY device_record FROM 'device_record.csv' CSV HEADER"
```