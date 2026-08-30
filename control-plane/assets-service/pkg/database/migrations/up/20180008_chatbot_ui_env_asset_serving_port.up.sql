-- add serving port to the existing environment preset of chatbot-ui
-- remove space in connection name, as it caused an invalid string (with space) in the workload spec

UPDATE assets_environment
SET connections = '[{"name": "ChatbotUI", "internalToolInfo": {"toolType": "chatbot-ui", "containerPort": 3000, "connectionType": "ExternalUrl", "externalUrlInfo": {}}}, {"name": "ServingPort", "internalToolInfo": {"toolType": "custom", "containerPort": 3000, "connectionType": "ServingPort", "servingPortInfo": {"protocol": "http"}}}]'
WHERE asset_id = '715a51c3-93ce-4838-a6c6-4d248e6b9ba0';
