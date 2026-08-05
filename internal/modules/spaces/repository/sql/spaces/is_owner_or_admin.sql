SELECT EXISTS (
    SELECT 1
    FROM space_members
    WHERE space_id = $1
      AND user_id = $2
      AND status = 'active'
      AND role IN ('owner', 'admin')
);
