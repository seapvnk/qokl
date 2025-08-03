module.exports = {
  branches: ['main'],
  plugins: [
    '@semantic-release/commit-analyzer',
    '@semantic-release/release-notes-generator',
    [
      '@semantic-release/exec',
      {
        publishCmd: 'bash build-release.sh ${nextRelease.version}',
      },
    ],
    [
      '@semantic-release/github',
      {
        assets: [
          'qokl-*.tar.gz',
          { 'path': 'qokl-linux-amd64-${nextRelease.version}.tar.gz', 'label': 'linux 64 distribution' }
        ],
      },
    ],
  ],
};

