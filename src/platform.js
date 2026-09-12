const TARGETS = {
  linux: {
    archiveFormat: "tar.gz",
    releaseOS: "linux",
  },
  darwin: {
    archiveFormat: "tar.gz",
    releaseOS: "darwin",
  },
  win32: {
    archiveFormat: "zip",
    releaseOS: "windows",
  },
};

function targetFor(platform = process.platform, architecture = process.arch) {
  const target = TARGETS[platform];
  const releaseArch = { x64: "amd64", arm64: "arm64" }[architecture];
  if (!target || !releaseArch || (platform === "win32" && releaseArch === "arm64")) {
    throw new Error(`unsupported runner platform or architecture: ${platform}/${architecture}`);
  }
  return { ...target, releaseArch };
}

function assetName(version, target) {
  return `ban-code-comments_${version}_${target.releaseOS}_${target.releaseArch}.${target.archiveFormat}`;
}

function releaseURLs(version, target) {
  const tag = `v${version}`;
  const base = `https://github.com/JustinDFuller/ban-code-comments/releases/download/${tag}`;
  return {
    archive: `${base}/${assetName(version, target)}`,
    checksums: `${base}/checksums.txt`,
  };
}

export { assetName, releaseURLs, targetFor };
