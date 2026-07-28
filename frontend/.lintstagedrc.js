module.exports = {
  // 返回数组时 lint-staged 不会再把文件路径拼到命令后，避免 Windows「参数太长」
  // Biome 自身用 --staged 读取暂存文件
  '**/*.{js,jsx,tsx,ts,css,less,md,cjs,mjs,json}': () => [
    'npx @biomejs/biome check --write --staged --files-ignore-unknown=true --no-errors-on-unmatched',
  ],
};
