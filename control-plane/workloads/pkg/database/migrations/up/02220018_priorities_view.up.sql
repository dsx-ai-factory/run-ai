DROP VIEW IF EXISTS priorities_cluster_view;

CREATE VIEW priorities_cluster_view AS
SELECT
    p.*,
    ps.id AS status_id,
    ps.cluster_id,
    ps.phase
FROM
    priorities as p
        LEFT JOIN priority_statuses as ps ON ps.priority_id = p.id;