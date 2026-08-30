-- Kept out of the main migrations directory on purpose: this migration only
-- exists to make the migrator fail, so the failure-recording path can be tested.
SELECT * FROM table_that_does_not_exist;
