module.exports = {
  branches: ['main'],
  plugins: [
    '@semantic-release/commit-analyzer',
    '@semantic-release/release-notes-generator',
    [
      '@semantic-release/github',
      {
        assets: ['qokl-*.tar.gz'],
      },
    ],
    [
      '@semantic-release/exec',
      {
        publishCmd: './build-release.sh ${nextRelease.version}',
      },
    ],
  ],
};

