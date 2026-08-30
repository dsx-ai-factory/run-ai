ALTER TABLE cluster_histories DROP CONSTRAINT cluster_histories_cluster_id_fkey;
ALTER TABLE cluster_histories ADD CONSTRAINT cluster_histories_cluster_id_fkey FOREIGN KEY (cluster_id) REFERENCES clusters(id) ON DELETE CASCADE;
