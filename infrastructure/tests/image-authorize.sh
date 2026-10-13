#!/bin/sh
# LB10b: proves the arbitrary-image serve path end to end through the Compose edge — an admin
# uploads an image to the API, nginx serves it only to a subject the image module authorizes
# (owner/member), with the correct extension-derived Content-Type, and denies everyone else by the
# /internal/image/authorize auth_request. Mirrors infrastructure/tests/playback-authorize.sh.
set -eu

run_hex="$(od -An -N6 -tx1 /dev/urandom | tr -d ' \n')"
course_id="10000000-0000-4000-8000-$run_hex"
module_id="61000000-0000-4000-8001-$run_hex"
lesson_id="61000000-0000-4000-8002-$run_hex"
admin_id="41000000-0000-4000-8004-$run_hex"
admin_email="image-$run_hex@courseforge.local"
outsider_id="41000000-0000-4000-8005-$run_hex"
outsider_email="image-outsider-$run_hex@courseforge.local"
http_port="${HTTP_PORT:-8080}"
origin="http://127.0.0.1:${http_port}"
cookie_jar="$(mktemp /tmp/courses-img-cookie.XXXXXX)"
outsider_cookie_jar="$(mktemp /tmp/courses-img-outsider.XXXXXX)"
png_file="$(mktemp /tmp/courses-img.XXXXXX.png)"
chmod 600 "$cookie_jar" "$outsider_cookie_jar"
# A real 1x1 PNG: DetectContentType sniffs the magic bytes and image.DecodeConfig reads 1x1.
printf '%s' 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+M8AAAICAQB3V6a2AAAAAElFTkSuQmCC' | base64 -d > "$png_file"

restore() {
  set +e
  image_row="$(docker compose exec -T postgres sh -c 'psql -tAq -U "$POSTGRES_USER" -d "$POSTGRES_DB"' <<SQL 2>/dev/null
SELECT id FROM lesson_images WHERE lesson_id='$lesson_id';
SQL
)"
  for id in $image_row; do
    docker compose exec -T api sh -c 'rm -f "/media/i/$1".* ' sh "$id" >/dev/null 2>&1
  done
  docker compose exec -T postgres sh -c 'psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB"' >/dev/null 2>&1 <<SQL
DELETE FROM lesson_images WHERE lesson_id='$lesson_id';
DELETE FROM course_lessons WHERE id='$lesson_id';
DELETE FROM course_modules WHERE id='$module_id';
DELETE FROM course_grants WHERE course_id='$course_id';
DELETE FROM courses WHERE id='$course_id';
DELETE FROM users WHERE id IN ('$admin_id','$outsider_id');
SQL
  rm -f "$cookie_jar" "$outsider_cookie_jar" "$png_file"
}
trap restore EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

assert_status() {
  name="$1"; expected="$2"; shift 2
  actual="$(curl --path-as-is -sS -o /dev/null -w '%{http_code}' "$@")"
  if [ "$actual" != "$expected" ]; then
    echo "$name: expected $expected, got $actual" >&2
    exit 1
  fi
  printf '%s status=%s\n' "$name" "$actual"
}

docker compose exec -T nginx nginx -t

# argon2id hash of 'courseforge-admin-demo' (same seed hash playback-authorize.sh uses).
hash='$argon2id$v=19$m=65536,t=3,p=2$YWRtaW4tZGVtby1zZWVkIQ$obXgLD59DMQsNm2cOgKWcl050QIrNheSnkrYE3HUvT0'
docker compose exec -T postgres sh -c 'psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB"' <<SQL
INSERT INTO users(id,email,display_name,role,password_hash,email_verified)
VALUES ('$admin_id','$admin_email','Image admin','admin','$hash',true);
INSERT INTO users(id,email,display_name,role,password_hash,email_verified)
VALUES ('$outsider_id','$outsider_email','Image outsider','student','$hash',true);
INSERT INTO courses(id,instructor_id,slug,title,summary,description,status,published_at)
VALUES ('$course_id','$admin_id','image-$run_hex','Image authorize','Image summary','Image description','published',now());
INSERT INTO course_modules(id,course_id,title,position) VALUES ('$module_id','$course_id','Module',0);
INSERT INTO course_lessons(id,module_id,title,position,is_preview) VALUES ('$lesson_id','$module_id','Lesson',0,false);
SQL

curl -fsS -o /dev/null -c "$cookie_jar" -X POST \
  -H 'Content-Type: application/json' -H "Origin: http://localhost:${http_port}" \
  --data "{\"email\":\"$admin_email\",\"password\":\"courseforge-admin-demo\"}" \
  "$origin/api/v1/auth/login"
curl -fsS -o /dev/null -c "$outsider_cookie_jar" -X POST \
  -H 'Content-Type: application/json' -H "Origin: http://localhost:${http_port}" \
  --data "{\"email\":\"$outsider_email\",\"password\":\"courseforge-admin-demo\"}" \
  "$origin/api/v1/auth/login"

upload_response="$(curl -fsS -b "$cookie_jar" -X POST \
  -H "Origin: http://localhost:${http_port}" \
  -F "file=@$png_file;type=application/octet-stream;filename=x.png" \
  "$origin/api/v1/admin/lessons/$lesson_id/images")"
image_id="$(printf '%s' "$upload_response" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')"
content_type="$(printf '%s' "$upload_response" | sed -n 's/.*"contentType":"\([^"]*\)".*/\1/p')"
if [ -z "$image_id" ] || [ "$content_type" != image/png ]; then
  echo "LB10b-upload: unexpected response: $upload_response" >&2
  exit 1
fi
printf 'LB10b-upload id=%s contentType=%s\n' "$image_id" "$content_type"

# The bytes landed where nginx serves them, named with the extension.
docker compose exec -T api sh -c 'test -f "/media/i/$1.png"' sh "$image_id"
printf 'LB10b-storage file=/media/i/%s.png present\n' "$image_id"

image_path="$origin/media/i/$image_id.png"
served_type="$(curl -sS -o /dev/null -b "$cookie_jar" -w '%{content_type}' "$image_path")"
assert_status LB10b-owner 200 -b "$cookie_jar" "$image_path"
case "$served_type" in
  image/png*) printf 'LB10b-owner content_type=%s\n' "$served_type" ;;
  *) echo "LB10b-owner: expected image/png Content-Type, got '$served_type'" >&2; exit 1 ;;
esac
assert_status LB10b-anonymous 403 "$image_path"
assert_status LB10b-outsider 403 -b "$outsider_cookie_jar" "$image_path"
assert_status LB10b-malformed 403 -b "$cookie_jar" "$origin/media/i/not-a-uuid.png"
assert_status LB10b-missing 403 -b "$cookie_jar" "$origin/media/i/51000000-0000-4000-8009-$run_hex.png"
# A path under /media that is not an image key stays refused by the catch-all.
assert_status LB10b-closed 403 -b "$cookie_jar" "$origin/media/i/$image_id"

printf '%s\n' 'PASS LB10b owner=200 anonymous=403 outsider=403 malformed=403 missing=403 closed=403'
