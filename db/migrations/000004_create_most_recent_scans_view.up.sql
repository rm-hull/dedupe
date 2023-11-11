CREATE OR REPLACE VIEW most_recent_scans_view AS
SELECT DISTINCT ON (hostname, root_directory) 
	id, 
	created_at,
	hostname, 
	root_directory, 
	scan_status
FROM scan
ORDER BY hostname, root_directory, created_at DESC;
