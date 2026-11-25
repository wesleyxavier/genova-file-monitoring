package analyzer

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	DefaultLlamaPort = 11440
	MinLlamaPort     = 11440
	MaxLlamaPort     = 11450
	ForbiddenPort    = 11435
	DefaultWebPort   = 18080
)

// EnvConfig agrupa as configurações carregadas do ambiente local (.env).
type EnvConfig struct {
	FalkorDBAddr     string
	FalkorDBUser     string
	FalkorDBPassword string
	LlamaServerCmd   string
	LlamaPort        int
	WebPort          int
}

// LoadEnvFile carrega pares KEY=VALUE de um arquivo .env sem sobrescrever variáveis já definidas.
func LoadEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.IndexByte(line, '=')
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		val = strings.Trim(val, `"'`)
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
	return scanner.Err()
}

// LoadEnvConfig carrega os arquivos .env informados (se existirem) e valida as configurações de execução.
func LoadEnvConfig(envPaths ...string) (EnvConfig, error) {
	for _, envPath := range envPaths {
		if envPath == "" {
			continue
		}
		if err := LoadEnvFile(envPath); err != nil {
			return EnvConfig{}, fmt.Errorf("erro ao ler %s: %w", envPath, err)
		}
	}

	addr := strings.TrimSpace(os.Getenv("FALKORDB_ADDR"))
	if addr == "" {
		addr = "127.0.0.1:6379"
	}

	llamaCmd := strings.TrimSpace(os.Getenv("LLAMA_SERVER_CMD"))
	llamaPort, err := ParseAndValidateLlamaPort(llamaCmd)
	if err != nil {
		return EnvConfig{}, err
	}

	webPort := DefaultWebPort
	if rawWebPort := strings.TrimSpace(os.Getenv("WEB_PORT")); rawWebPort != "" {
		p, err := strconv.Atoi(rawWebPort)
		if err != nil || p < 1 || p > 65535 {
			return EnvConfig{}, fmt.Errorf("WEB_PORT inválida: %q", rawWebPort)
		}
		webPort = p
	}

	return EnvConfig{
		FalkorDBAddr:     addr,
		FalkorDBUser:     os.Getenv("FALKORDB_USER"),
		FalkorDBPassword: os.Getenv("FALKORDB_PASSWORD"),
		LlamaServerCmd:   llamaCmd,
		LlamaPort:        llamaPort,
		WebPort:          webPort,
	}, nil
}

// ParseAndValidateLlamaPort extrai a porta do comando do llama-server e valida a faixa 11440-11450 (nunca 11435).
func ParseAndValidateLlamaPort(cmdStr string) (int, error) {
	port := DefaultLlamaPort
	if cmdStr != "" {
		fields := strings.Fields(cmdStr)
		for i := 0; i < len(fields); i++ {
			arg := fields[i]
			if (arg == "--port" || arg == "-p") && i+1 < len(fields) {
				p, err := strconv.Atoi(fields[i+1])
				if err != nil {
					return 0, fmt.Errorf("porta do llama-server inválida: %q", fields[i+1])
				}
				port = p
				break
			}
			if strings.HasPrefix(arg, "--port=") {
				raw := strings.TrimPrefix(arg, "--port=")
				p, err := strconv.Atoi(raw)
				if err != nil {
					return 0, fmt.Errorf("porta do llama-server inválida: %q", raw)
				}
				port = p
				break
			}
		}
	}

	if err := ValidateLlamaPort(port); err != nil {
		return 0, err
	}
	return port, nil
}

// ValidateLlamaPort garante que a porta esteja na faixa 11440-11450 e nunca seja 11435.
func ValidateLlamaPort(port int) error {
	if port == ForbiddenPort {
		return fmt.Errorf("porta %d é proibida pela constituição", ForbiddenPort)
	}
	if port < MinLlamaPort || port > MaxLlamaPort {
		return fmt.Errorf("porta do llama-server (%d) fora da faixa permitida %d-%d", port, MinLlamaPort, MaxLlamaPort)
	}
	return nil
}
