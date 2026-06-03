BEGIN;

DROP TRIGGER IF EXISTS trg_notifications_touch_updated_at ON notifications;
DROP TRIGGER IF EXISTS trg_request_response_artifacts_touch_updated_at ON request_response_artifacts;
DROP TRIGGER IF EXISTS trg_contribution_artifacts_touch_updated_at ON contribution_artifacts;
DROP TRIGGER IF EXISTS trg_space_artifacts_touch_updated_at ON space_artifacts;
DROP TRIGGER IF EXISTS trg_profile_artifacts_touch_updated_at ON profile_artifacts;
DROP TRIGGER IF EXISTS trg_contributions_touch_updated_at ON contributions;
DROP TRIGGER IF EXISTS trg_artifacts_touch_updated_at ON artifacts;
DROP TRIGGER IF EXISTS trg_request_responses_touch_updated_at ON request_responses;
DROP TRIGGER IF EXISTS trg_request_skills_touch_updated_at ON request_skills;
DROP TRIGGER IF EXISTS trg_requests_touch_updated_at ON requests;
DROP TRIGGER IF EXISTS trg_space_skills_touch_updated_at ON space_skills;
DROP TRIGGER IF EXISTS trg_space_members_touch_updated_at ON space_members;
DROP TRIGGER IF EXISTS trg_spaces_touch_updated_at ON spaces;
DROP TRIGGER IF EXISTS trg_user_skills_touch_updated_at ON user_skills;
DROP TRIGGER IF EXISTS trg_skills_touch_updated_at ON skills;
DROP TRIGGER IF EXISTS trg_parsed_resume_drafts_touch_updated_at ON parsed_resume_drafts;
DROP TRIGGER IF EXISTS trg_uploaded_resumes_touch_updated_at ON uploaded_resumes;
DROP TRIGGER IF EXISTS trg_profile_links_touch_updated_at ON profile_links;
DROP TRIGGER IF EXISTS trg_profiles_touch_updated_at ON profiles;
DROP TRIGGER IF EXISTS trg_auth_identities_touch_updated_at ON auth_identities;
DROP TRIGGER IF EXISTS trg_users_touch_updated_at ON users;

COMMIT;
