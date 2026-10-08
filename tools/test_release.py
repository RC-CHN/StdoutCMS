"""Exercise release tooling in disposable repositories, without a remote."""
import pathlib
import shutil
import subprocess
import tempfile
import unittest


class ReleaseTests(unittest.TestCase):
    def setUp(self):
        temporary = tempfile.TemporaryDirectory(prefix="stdoutcms-release-")
        self.addCleanup(temporary.cleanup)
        self.root = pathlib.Path(temporary.name)
        (self.root / "tools").mkdir()
        for script in ("release.sh", "check-release.sh"):
            shutil.copyfile(pathlib.Path(__file__).with_name(script), self.root / "tools" / script)
        self.run_command("git", "init", "-b", "main")
        self.git("config", "user.name", "Release Test")
        self.git("config", "user.email", "release@example.invalid")
        self.git("config", "commit.gpgsign", "false")
        self.git("config", "tag.gpgsign", "false")
        self.write("VERSION", "0.3.0\n")
        self.commit("chore: set up release tooling")

    def run_command(self, *args, success=True):
        result = subprocess.run(args, cwd=self.root, capture_output=True, text=True)
        if success:
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        else:
            self.assertNotEqual(result.returncode, 0, result.stdout + result.stderr)
        return result.stdout.strip()

    def git(self, *args):
        return self.run_command("git", *args)

    def write(self, filename, content):
        (self.root / filename).write_text(content)

    def commit(self, message):
        self.git("add", ".")
        self.git("commit", "-m", message)

    def release(self, version="0.4.0", **kwargs):
        return self.run_command("sh", "tools/release.sh", version, **kwargs)

    def check(self, tag="v0.4.0", **kwargs):
        return self.run_command("sh", "tools/check-release.sh", tag, **kwargs)

    def test_release_creates_dedicated_commit_and_annotated_tag(self):
        self.release()
        self.assertEqual(self.git("log", "-1", "--format=%s"), "chore: release v0.4.0")
        self.assertEqual(self.git("cat-file", "-t", "refs/tags/v0.4.0"), "tag")
        self.assertEqual(self.git("rev-parse", "v0.4.0^{commit}"), self.git("rev-parse", "HEAD"))
        self.assertEqual(self.git("diff-tree", "--no-commit-id", "--name-only", "-r", "HEAD"), "VERSION")
        self.assertEqual(self.git("status", "--porcelain"), "")
        # Validation must inspect the tagged tree, even if the working copy changes.
        self.write("VERSION", "9.9.9\n")
        self.check()

    def test_invalid_or_nonincreasing_version_leaves_repository_untouched(self):
        head = self.git("rev-parse", "HEAD")
        for version in ("v0.4.0", "0.4", "01.4.0", "0.4.0-rc.1", "0.3.0", "0.2.0"):
            with self.subTest(version=version):
                self.release(version, success=False)
                self.assertEqual(self.git("rev-parse", "HEAD"), head)
                self.assertEqual(self.git("status", "--porcelain"), "")

    def test_dirty_tree_is_rejected(self):
        self.write("pending.txt", "unfinished")
        self.release(success=False)
        self.assertEqual((self.root / "VERSION").read_text(), "0.3.0\n")

    def test_existing_tag_is_rejected(self):
        self.git("tag", "v0.4.0")
        self.release(success=False)
        self.assertEqual(self.git("status", "--porcelain"), "")

    def test_non_main_branch_is_rejected(self):
        self.git("checkout", "-b", "feature")
        self.release(success=False)

    def test_tag_on_feature_commit_is_rejected(self):
        self.write("VERSION", "0.4.0\n")
        self.commit("feat: unrelated change")
        self.git("tag", "v0.4.0")
        self.check(success=False)

    def test_mismatched_tag_is_rejected(self):
        self.release()
        self.git("tag", "v0.5.0")
        self.check("v0.5.0", success=False)

    def test_release_commit_containing_other_changes_is_rejected(self):
        self.write("VERSION", "0.4.0\n")
        self.write("feature.txt", "should be committed separately")
        self.commit("chore: release v0.4.0")
        self.git("tag", "v0.4.0")
        self.check(success=False)


if __name__ == "__main__":
    unittest.main()
