create table stake_remittance
(
    id                 bigint auto_increment
        primary key,
    bet_id             bigint                              not null,
    status             varchar(50)                         null,
    status_description int(100)                            null,
    created            datetime  default CURRENT_TIMESTAMP not null,
    updated            timestamp default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP,
    constraint bet_id
        unique (bet_id)
);

create index created
    on stake_remittance (created);

create index status
    on stake_remittance (status);

create index status_description
    on stake_remittance (status_description);

create index updated
    on stake_remittance (updated);

