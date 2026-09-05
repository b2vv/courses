-- Local development identity fixtures. All rows commit together and reconcile to this file.
BEGIN;
LOCK TABLE users IN SHARE ROW EXCLUSIVE MODE;
INSERT INTO users (id, email, display_name, role, password_hash) VALUES
('40000000-0000-4000-8000-000000000001', 'student@courseforge.local', 'Студентка Курсу', 'student', '$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNsdA$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is'),
('40000000-0000-4000-8000-000000000002', 'instructor@courseforge.local', 'Викладач Курсу', 'instructor', '$argon2id$v=19$m=65536,t=3,p=2$aW5zdHJ1Y3Rvci1zZWVkIQ$zjeAhI2zv+SPciv5gZZegL3BFoWNrJGqMg721CkxRrI'),
('40000000-0000-4000-8000-000000000003', 'admin@courseforge.local', 'Адміністратор Курсу', 'admin', '$argon2id$v=19$m=65536,t=3,p=2$YWRtaW4tZGVtby1zZWVkIQ$obXgLD59DMQsNm2cOgKWcl050QIrNheSnkrYE3HUvT0'),
('40000000-0000-4000-8000-0000000000d1', 'block-me@courseforge.local', 'Тест блокування', 'student', '$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNsdA$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is'),
('40000000-0000-4000-8000-0000000000d2', 'delete-me@courseforge.local', 'Тест видалення', 'student', '$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNsdA$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is'),
('40000000-0000-4000-8000-0000000000d3', 'reset-me@courseforge.local', 'Тест скидання пароля', 'student', '$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNsdA$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is')
ON CONFLICT (id) DO UPDATE SET
    email = EXCLUDED.email,
    display_name = EXCLUDED.display_name,
    role = EXCLUDED.role,
    password_hash = EXCLUDED.password_hash,
    disabled_at = NULL,
    deleted_at = NULL,
    updated_at = now();
COMMIT;
