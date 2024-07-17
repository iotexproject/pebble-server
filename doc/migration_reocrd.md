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
| 2024-07-15 | 2024-07-15 01:50:00(migrated) | 351358813281760            | from envrioBLOQ                                   |
|            |                               | 351358813282131            |                                                   |
|            |                               | 350916067066608            |                                                   |
|            |                               | 351358813281992            |                                                   |
|            |                               | 351358813065718            |                                                   |
|            |                               | 351358813357594            |                                                   |
| 2024-07-17 | 2024-07-17 07:00:00(planed)   | 350916066753800            | missing blockchain event                          |
|            |                               | 350916066755219            |                                                   |
|            |                               | 350916067051147            |                                                   |
|            |                               | 350916067066178            |                                                   |
|            |                               | 350916067066269            |                                                   |
|            |                               | 350916067066673            |                                                   |
|            |                               | 350916067070824            |                                                   |
|            |                               | 350916067094295            |                                                   |
|            |                               | 351358810263407            |                                                   |
|            |                               | 351358813083174            |                                                   |
|            |                               | 351358813094361            |                                                   |
|            |                               | 351358813280705            |                                                   |
|            |                               | 351358813281182            |                                                   |
|            |                               | 351358813374102            |                                                   |
|            |                               | 351358815441396            |                                                   |
|            |                               | %0                         |                                                   |

migrate steps:

1. prepare wasm with blacklist(devices in '2024-07-11')
2. publish wasm and restart old w3bstream
3. modify new pebble sequencer white list and restart new pebble sequencer(tag
   v0.2.3)

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
PGPASSWORD=$NEW_PEBBLE_DB_PASSWORD \
psql -h $NEW_PEBBLE_DB -U $NEW_PEBBLE_DB_USER -d $NEW_PEBBLE_DB_NAME -c \
"\COPY (SELECT * FROM device where updated_at > '2024-07-11 08:30:00' and id in('350916067079072','350916067070162','350916067099906','350916067070345','351358810283462')) TO 'device_0711.csv' CSV HEADER"

PGPASSWORD=$NEW_PEBBLE_DB_PASSWORD \
psql -h $NEW_PEBBLE_DB -U $NEW_PEBBLE_DB_USER -d $NEW_PEBBLE_DB_NAME -c \
"\COPY (SELECT * FROM device where updated_at > '2024-07-15 01:50:00' and id in('351358813281760','351358813282131','350916067066608','351358813281992','351358813065718','351358813357594')) TO 'device_0715.csv' CSV HEADER"


PGPASSWORD=$NEW_PEBBLE_DB_PASSWORD \
psql -h $NEW_PEBBLE_DB -U $NEW_PEBBLE_DB_USER -d $NEW_PEBBLE_DB_NAME -c \
"\COPY (SELECT * FROM device_record where updated_at > '2024-07-11 08:30:00' and imei in('350916067079072','350916067070162','350916067099906','350916067070345','351358810283462')) TO 'device_record_0711.csv' CSV HEADER"

PGPASSWORD=$NEW_PEBBLE_DB_PASSWORD \
psql -h $NEW_PEBBLE_DB -U $NEW_PEBBLE_DB_USER -d $NEW_PEBBLE_DB_NAME -c \
"\COPY (SELECT * FROM device_record where updated_at > '2024-07-15 01:50:00' and id in('351358813281760','351358813282131','350916067066608','351358813281992','351358813065718','351358813357594')) TO 'device_record_0715.csv' CSV HEADER"
```

## import to 34.172.94.245(w3bstream pebble wasm db)

```shell
PGPASSWORD=$OLD_PEBBLE_DB_PASSWORD \
psql -h $OLD_PEBBLE_DB -U $OLD_PEBBLE_DB_USER -d $OLD_PEBBLE_DB_NAME -c \
"\COPY device FROM 'device_0711.csv' CSV HEADER"

PGPASSWORD=$OLD_PEBBLE_DB_PASSWORD \
psql -h $OLD_PEBBLE_DB -U $OLD_PEBBLE_DB_USER -d $OLD_PEBBLE_DB_NAME -c \
"\COPY device FROM 'device_0715.csv' CSV HEADER"

PGPASSWORD=$OLD_PEBBLE_DB_PASSWORD \
psql -h $OLD_PEBBLE_DB -U $OLD_PEBBLE_DB_USER -d $OLD_PEBBLE_DB_NAME -c \
"\COPY device_record FROM 'device_record_0711.csv' CSV HEADER"

PGPASSWORD=$OLD_PEBBLE_DB_PASSWORD \
psql -h $OLD_PEBBLE_DB -U $OLD_PEBBLE_DB_USER -d $OLD_PEBBLE_DB_NAME -c \
"\COPY device_record FROM 'device_record_0715.csv' CSV HEADER"
```

## diff

```shell
PGPASSWORD=$OLD_PEBBLE_DB_PASSWORD psql -h $OLD_PEBBLE_DB -U $OLD_PEBBLE_DB_USER -d $OLD_PEBBLE_DB_NAME -c "\
  COPY ( \
    SELECT (
      id,name,owner,address,status,avatar,config,real_firmware,total_gas,bulk_upload,data_channel,upload_period,bulk_upload_sampling_cnt,bulk_upload_sampling_freq,beep,state,type,configurable \
    ) \
    FROM \
      device \
    WHERE 
      id not in('103381234567407','351358810263647','351358813375182','350916067079072','350916067070162','350916067099906','350916067070345','351358810283462','351358813281760','351358813282131','350916067066608','351358813281992','351358813065718','351358813357594')
    ORDER BY \
      id \
  ) TO STDOUT WITH CSV HEADER" >device_old.csv

PGPASSWORD=$NEW_PEBBLE_DB_PASSWORD psql -h $NEW_PEBBLE_DB -U $NEW_PEBBLE_DB_USER -d $NEW_PEBBLE_DB_NAME -c "\
  COPY ( \
    SELECT (
      id,name,owner,address,status,avatar,config,real_firmware,total_gas,bulk_upload,data_channel,upload_period,bulk_upload_sampling_cnt,bulk_upload_sampling_freq,beep,state,type,configurable \
    ) \
    FROM \
      device \
    WHERE 
      id not in('103381234567407','351358810263647','351358813375182','350916067079072','350916067070162','350916067099906','350916067070345','351358810283462','351358813281760','351358813282131','350916067066608','351358813281992','351358813065718','351358813357594')
    ORDER BY \
      id \
  ) TO STDOUT WITH CSV HEADER" >device_new.csv

```