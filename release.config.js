module.exports = {
    branches: ['master'],
    plugins: [
        "@semantic-release/commit-analyzer",
        "@semantic-release/release-notes-generator",
        ["@semantic-release/github", {
            "assets": [
                {
                    "path": "release/win32.zip",
                    "label": "Windows (32bit)",
                    "name": "dockerized-${nextRelease.gitTag}-win32.zip"
                },
                {
                    "path": "release/win64.zip",
                    "label": "Windows (64bit)",
                    "name": "dockerized-${nextRelease.gitTag}-win64.zip"
                },
                {
                    "path": "release/linux-x86_64.zip",
                    "label": "Linux (x86/64)",
                    "name": "dockerized-${nextRelease.gitTag}-linux-x86_64.zip"
                },
                {
                    "path": "release/mac-x86_64.zip",
                    "label": "MacOS (Intel)",
                    "name": "dockerized-${nextRelease.gitTag}-mac-x86_64.zip"
                },
                {
                    "path": "release/mac-arm64.zip",
                    "label": "MacOS (Apple Silicon)",
                    "name": "dockerized-${nextRelease.gitTag}-mac-arm64.zip"
                },
            ]
        }],
        '@semantic-release/changelog',
        '@semantic-release/git',
    ],
};
