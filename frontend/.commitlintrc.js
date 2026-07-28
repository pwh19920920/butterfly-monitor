module.exports = {
  extends: ['@commitlint/config-conventional'],
  rules: {
    // 允许不写 type / subject，支持中文自由提交
    'type-empty': [0],
    'subject-empty': [0],
    // 主题不做大小写限制（中文不受影响，英文也更宽松）
    'subject-case': [0],
    // 允许任意 type（含中文整句）
    'type-enum': [0],
  },
};
