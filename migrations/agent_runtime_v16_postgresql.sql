-- Agent Runtime v16: persist rich tool results in chat messages
-- Adds toolresults JSONB column to chatmessages to store ui_render, html_report, etc.

ALTER TABLE chatmessages
    ADD COLUMN IF NOT EXISTS toolresults JSONB NULL;

COMMENT ON COLUMN chatmessages.toolresults IS
    'JSON array of rich tool results produced during an agent run, e.g. [{tool, result}]. Stored as {"items":[...]}';
