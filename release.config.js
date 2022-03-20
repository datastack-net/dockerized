module.exports = {
  branches: ['master', 'feature/semantic-release'],
  plugins: [
    "@semantic-release/commit-analyzer",
    "@semantic-release/release-notes-generator",
    "@semantic-release/github",
    '@semantic-release/changelog',
    '@semantic-release/git',
  ],
};
