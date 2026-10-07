-- Explicit demo seed. All rows commit together and repeated/concurrent runs are safe.
BEGIN;
LOCK TABLE courses IN SHARE ROW EXCLUSIVE MODE;
INSERT INTO courses (id, instructor_id, slug, title, summary, description, status, created_at, updated_at, published_at) VALUES
('10000000-0000-4000-8000-000000000001', '40000000-0000-4000-8000-000000000002', 'go-backend-foundations', 'Go: фундамент надійного бекенду', 'Побудуйте HTTP API з чіткими межами домену, PostgreSQL і перевірками.', 'Практичний курс зі створення Go API. Розглянемо модульну архітектуру, контракти, транзакції PostgreSQL, обробку помилок і тести. Каталог демонструє структуру платформи; відеоуроки й оплата поки недоступні.', 'published', '2026-08-01T10:00:00Z', '2026-09-03T10:00:00Z', '2026-09-03T10:00:00Z'),
('10000000-0000-4000-8000-000000000002', '40000000-0000-4000-8000-000000000002', 'react-accessible-interfaces', 'React: доступні інтерфейси', 'Створюйте зрозумілі інтерфейси з React, TypeScript і клавіатурною навігацією.', 'Від семантичного HTML до стану застосунку: компоненти, форми, навігація та доступність. Працюємо з явними контрактами й перевіряємо поведінку користувача. Це демонстраційна сторінка курсу foundation-версії.', 'published', '2026-08-01T10:00:00Z', '2026-09-02T10:00:00Z', '2026-09-02T10:00:00Z'),
('10000000-0000-4000-8000-000000000003', '40000000-0000-4000-8000-000000000002', 'system-design-practice', 'Системний дизайн на практиці', 'Навчіться обирати прості архітектурні рішення й перевіряти їхні обмеження.', 'Проєктування починається з потреб користувача. У курсі розбираємо межі модулів, узгодженість даних, черги, спостережуваність і розвиток системи. Навчальні матеріали будуть доступні в наступних етапах платформи.', 'published', '2026-08-01T10:00:00Z', '2026-09-01T10:00:00Z', '2026-09-01T10:00:00Z'),
('10000000-0000-4000-8000-000000000004', '40000000-0000-4000-8000-000000000002', 'draft-course', 'Чернетка курсу', 'Приклад неопублікованого курсу.', 'Цей запис перевіряє приховування чернеток у публічному каталозі.', 'draft', '2026-08-01T10:00:00Z', '2026-08-01T10:00:00Z', NULL),
('10000000-0000-4000-8000-000000000005', '40000000-0000-4000-8000-000000000002', 'archived-course', 'Архівний курс', 'Приклад архівного курсу.', 'Цей запис перевіряє приховування архівних курсів у публічному каталозі.', 'archived', '2026-08-01T10:00:00Z', '2026-09-01T10:00:00Z', '2026-08-02T10:00:00Z')
ON CONFLICT (id) DO UPDATE SET
    instructor_id = EXCLUDED.instructor_id,
    slug = EXCLUDED.slug,
    title = EXCLUDED.title,
    summary = EXCLUDED.summary,
    description = EXCLUDED.description,
    status = EXCLUDED.status,
    created_at = EXCLUDED.created_at,
    updated_at = EXCLUDED.updated_at,
    published_at = EXCLUDED.published_at;

-- The student fixture is an active self-member of every published free demo course, so the
-- enrollment access rule does not make the existing authenticated learning journey disappear.
-- Guarded by to_regclass so seed.sql stays loadable against a pre-00011 schema: the course
-- migration-preservation tests (course/adapters/postgres) run this whole seed after DownTo an
-- earlier version, where course_grants does not yet exist. Dynamic EXECUTE is never planned
-- when the branch is skipped, so the missing relation cannot raise 42P01.
DO $$
BEGIN
  IF to_regclass('public.course_grants') IS NOT NULL THEN
    EXECUTE $grant$
      INSERT INTO course_grants (id, user_id, course_id, source, order_id, state, granted_by) VALUES
      ('30000000-0000-4000-8000-000000000001', '40000000-0000-4000-8000-000000000001', '10000000-0000-4000-8000-000000000001', 'self', NULL, 'active', NULL),
      ('30000000-0000-4000-8000-000000000002', '40000000-0000-4000-8000-000000000001', '10000000-0000-4000-8000-000000000002', 'self', NULL, 'active', NULL),
      ('30000000-0000-4000-8000-000000000003', '40000000-0000-4000-8000-000000000001', '10000000-0000-4000-8000-000000000003', 'self', NULL, 'active', NULL)
      ON CONFLICT (user_id, course_id) DO NOTHING
    $grant$;
  END IF;
END $$;
COMMIT;
