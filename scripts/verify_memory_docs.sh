#!/bin/bash

# GreenSoulAI 记忆系统文档验证脚本
# 用途：确保设计思想文档与实际代码实现保持一致

echo "🔍 GreenSoulAI 记忆系统文档验证"
echo "================================="

# 设置项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# 验证结果统计
PASS_COUNT=0
FAIL_COUNT=0

# 验证函数
verify_step() {
    local description="$1"
    local condition="$2"
    
    echo -n "检查 $description ... "
    
    if eval "$condition"; then
        echo "✅ 通过"
        ((PASS_COUNT++))
    else
        echo "❌ 失败"
        ((FAIL_COUNT++))
    fi
}

echo "📁 1. 核心文件存在性验证"
echo "-------------------------"

verify_step "ContextualMemory实现文件" \
    "[ -f 'internal/memory/contextual/contextual_memory.go' ]"

verify_step "MemoryManager实现文件" \
    "[ -f 'internal/crew/memory_manager.go' ]"

verify_step "基础Memory接口文件" \
    "[ -f 'internal/memory/memory.go' ]"

verify_step "记忆系统完整指南" \
    "[ -f 'docs/MEMORY_SYSTEM_GUIDE.md' ]"

verify_step "验证机制文档" \
    "[ -f 'docs/MEMORY_DESIGN_VERIFICATION.md' ]"

echo ""
echo "🔧 2. 核心结构体定义验证"
echo "-------------------------"

verify_step "ContextualMemory结构体定义" \
    "grep -q 'type ContextualMemory struct' internal/memory/contextual/contextual_memory.go"

verify_step "MemoryManager结构体定义" \
    "grep -q 'type MemoryManager struct' internal/crew/memory_manager.go"

verify_step "Memory接口定义" \
    "grep -q 'type Memory interface' internal/memory/memory.go"

verify_step "MemoryItem数据结构定义" \
    "grep -q 'type MemoryItem struct' internal/memory/memory.go"

echo ""
echo "🎯 3. 关键方法实现验证"
echo "-------------------------"

verify_step "BuildContextForTask方法存在" \
    "grep -q 'func.*BuildContextForTask' internal/memory/contextual/contextual_memory.go"

verify_step "MemoryManager的BuildTaskContext方法" \
    "grep -q 'func.*BuildTaskContext' internal/crew/memory_manager.go"

verify_step "fetchSTMContext方法实现" \
    "grep -q 'func.*fetchSTMContext' internal/memory/contextual/contextual_memory.go"

verify_step "fetchLTMContext方法实现" \
    "grep -q 'func.*fetchLTMContext' internal/memory/contextual/contextual_memory.go"

echo ""
echo "📊 4. 配置结构验证"
echo "-------------------------"

verify_step "ContextualMemoryConfig定义" \
    "grep -q 'type ContextualMemoryConfig struct' internal/memory/contextual/contextual_memory.go"

verify_step "MemoryManagerConfig定义" \
    "grep -q 'type MemoryManagerConfig struct' internal/crew/memory_manager.go"

verify_step "EmbedderConfig定义" \
    "grep -q 'type EmbedderConfig struct' internal/memory/memory.go"

echo ""
echo "🧪 5. 示例代码功能验证"
echo "-------------------------"

verify_step "记忆到LLM示例存在" \
    "[ -f 'examples/memory/memory_to_llm_example.go' ]"

# 编译检查
echo -n "检查示例代码编译正确性 ... "
if go build -o /tmp/memory_test examples/memory/memory_to_llm_example.go > /dev/null 2>&1; then
    echo "✅ 通过"
    ((PASS_COUNT++))
    rm -f /tmp/memory_test
else
    echo "❌ 失败"
    ((FAIL_COUNT++))
fi

echo ""
echo "📚 6. 文档内容一致性验证"
echo "-------------------------"

# 检查文档中是否包含了关键的技术概念
verify_step "文档包含认知科学基础描述" \
    "grep -q '认知科学基础' docs/MEMORY_SYSTEM_GUIDE.md"

verify_step "文档包含ContextualMemory技术细节" \
    "grep -q 'ContextualMemory' docs/MEMORY_SYSTEM_GUIDE.md"

verify_step "文档包含并行检索描述" \
    "grep -q '并行检索' docs/MEMORY_SYSTEM_GUIDE.md"

verify_step "文档包含数据流转描述" \
    "grep -q '数据流转机制' docs/MEMORY_SYSTEM_GUIDE.md"

verify_step "文档包含性能优势描述" \
    "grep -q '性能与优势' docs/MEMORY_SYSTEM_GUIDE.md"

echo ""
echo "🔄 7. 实际功能测试"
echo "-------------------------"

# 创建临时测试目录
TEST_DIR="/tmp/greensoul_memory_test"
mkdir -p "$TEST_DIR"

echo -n "检查记忆系统基础功能 ... "
if cd examples/memory && go run memory_to_llm_example.go > "$TEST_DIR/test_output.log" 2>&1; then
    if grep -q "演示完成" "$TEST_DIR/test_output.log" && grep -q "数据流转统计" "$TEST_DIR/test_output.log"; then
        echo "✅ 通过"
        ((PASS_COUNT++))
    else
        echo "❌ 失败（功能异常）"
        ((FAIL_COUNT++))
    fi
else
    echo "❌ 失败（执行异常）"
    ((FAIL_COUNT++))
fi

cd "$PROJECT_ROOT"

echo ""
echo "📊 验证结果统计"
echo "================"
echo "✅ 通过测试: $PASS_COUNT 项"
echo "❌ 失败测试: $FAIL_COUNT 项"
echo "🎯 成功率: $(( PASS_COUNT * 100 / (PASS_COUNT + FAIL_COUNT) ))%"

if [ $FAIL_COUNT -eq 0 ]; then
    echo ""
    echo "🎉 所有验证项目都通过了！"
    echo "📖 设计思想文档与代码实现保持高度一致"
    echo "✅ 文档正确性和结构清晰度得到确认"
    exit 0
else
    echo ""
    echo "⚠️  有 $FAIL_COUNT 项验证失败，请检查相关问题"
    echo "💡 建议："
    echo "   1. 检查失败的文件和方法是否存在"
    echo "   2. 确认代码编译无误"
    echo "   3. 验证文档内容是否需要更新"
    exit 1
fi
