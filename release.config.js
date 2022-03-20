module.exports = {
  branches: ['master', 'feature/semantic-release', 'feature/semantic-release-build'],
  plugins: [
    "@semantic-release/commit-analyzer",
    "@semantic-release/release-notes-generator",
    ["@semantic-release/github", {
      "assets": [
        {"path": "release/*.zip", "label": "Binary distribution"},
      ]
    }],
    '@semantic-release/changelog',
    '@semantic-release/git',
  ],
};
