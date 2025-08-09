# Garden Example - GreenSoulAI

## 运行方式

设置API密钥并运行：

```bash
export OPENROUTER_API_KEY="your-openrouter-api-key-here"
GARDEN_MODEL="moonshotai/kimi-k2:free" GARDEN_VERBOSE=0 go run ./examples/garden/cmd
```

或者一行命令：

```bash
OPENROUTER_API_KEY="your-api-key" GARDEN_MODEL="moonshotai/kimi-k2:free" GARDEN_VERBOSE=0 go run ./examples/garden/cmd
```

## 获取API密钥

访问 [OpenRouter](https://openrouter.ai/) 注册并获取免费API密钥。