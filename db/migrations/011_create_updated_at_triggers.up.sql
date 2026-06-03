BEGIN;

CREATE TRIGGER trg_users_touch_updated_at
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

CREATE TRIGGER trg_auth_identities_touch_updated_at
BEFORE UPDATE ON auth_identities
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

CREATE TRIGGER trg_profiles_touch_updated_at
BEFORE UPDATE ON profiles
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

CREATE TRIGGER trg_profile_links_touch_updated_at
BEFORE UPDATE ON profile_links
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

CREATE TRIGGER trg_uploaded_resumes_touch_updated_at
BEFORE UPDATE ON uploaded_resumes
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

CREATE TRIGGER trg_parsed_resume_drafts_touch_updated_at
BEFORE UPDATE ON parsed_resume_drafts
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

CREATE TRIGGER trg_skills_touch_updated_at
BEFORE UPDATE ON skills
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

CREATE TRIGGER trg_user_skills_touch_updated_at
BEFORE UPDATE ON user_skills
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

CREATE TRIGGER trg_spaces_touch_updated_at
BEFORE UPDATE ON spaces
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

CREATE TRIGGER trg_space_members_touch_updated_at
BEFORE UPDATE ON space_members
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

CREATE TRIGGER trg_space_skills_touch_updated_at
BEFORE UPDATE ON space_skills
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

CREATE TRIGGER trg_requests_touch_updated_at
BEFORE UPDATE ON requests
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

CREATE TRIGGER trg_request_skills_touch_updated_at
BEFORE UPDATE ON request_skills
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

CREATE TRIGGER trg_request_responses_touch_updated_at
BEFORE UPDATE ON request_responses
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

CREATE TRIGGER trg_artifacts_touch_updated_at
BEFORE UPDATE ON artifacts
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

CREATE TRIGGER trg_contributions_touch_updated_at
BEFORE UPDATE ON contributions
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

CREATE TRIGGER trg_profile_artifacts_touch_updated_at
BEFORE UPDATE ON profile_artifacts
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

CREATE TRIGGER trg_space_artifacts_touch_updated_at
BEFORE UPDATE ON space_artifacts
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

CREATE TRIGGER trg_contribution_artifacts_touch_updated_at
BEFORE UPDATE ON contribution_artifacts
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

CREATE TRIGGER trg_request_response_artifacts_touch_updated_at
BEFORE UPDATE ON request_response_artifacts
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

CREATE TRIGGER trg_notifications_touch_updated_at
BEFORE UPDATE ON notifications
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

COMMIT;
