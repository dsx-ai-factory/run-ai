alter table containers
drop constraint if exists fk_containers_unique_pods;

alter table containers
    add constraint fk_containers_unique_pods
        foreign key (pod_id) references pods;