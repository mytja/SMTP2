ALTER TABLE attachments ADD COLUMN url_to_forward VARCHAR(300) DEFAULT "";
ALTER TABLE attachments ADD COLUMN is_forwarded BOOLEAN DEFAULT false;
