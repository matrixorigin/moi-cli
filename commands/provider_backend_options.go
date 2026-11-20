package commands

import (
	"flag"
	"os"
	"strconv"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

type providerBackendCLIOptions struct {
	name, providerType, reasoningProtocol  string
	apiKeyFile, address, models, mimeTypes string
	timeoutSeconds                         int32
}

func providerBackendInt32Value(target *int32) func(string) error {
	return func(raw string) error {
		value, err := strconv.ParseInt(raw, 10, 32)
		if err != nil {
			return err
		}
		*target = int32(value)
		return nil
	}
}

func addProviderBackendCreateFlags(fs *flag.FlagSet) *providerBackendCLIOptions {
	options := &providerBackendCLIOptions{}
	fs.StringVar(&options.providerType, "type", "", "Provider type")
	fs.StringVar(&options.reasoningProtocol, "reasoning-control-protocol", "", "Reasoning control protocol")
	fs.StringVar(&options.apiKeyFile, "api-key-file", "", "Provider API key file (read only when supplied)")
	fs.StringVar(&options.address, "address", "", "Provider address")
	fs.Func("timeout-seconds", "Provider timeout in seconds", providerBackendInt32Value(&options.timeoutSeconds))
	fs.StringVar(&options.models, "models", "", "Provider models (comma-separated)")
	fs.StringVar(&options.mimeTypes, "supported-mime-types", "", "Parser supported MIME types (comma-separated)")
	return options
}

func addProviderBackendUpdateFlags(fs *flag.FlagSet) *providerBackendCLIOptions {
	options := &providerBackendCLIOptions{}
	fs.StringVar(&options.name, "name", "", "Updated backend name")
	fs.StringVar(&options.apiKeyFile, "api-key-file", "", "Provider API key file (empty file sends an explicit empty key)")
	fs.StringVar(&options.reasoningProtocol, "reasoning-control-protocol", "", "Reasoning control protocol")
	fs.Func("timeout-seconds", "Provider timeout in seconds", providerBackendInt32Value(&options.timeoutSeconds))
	fs.StringVar(&options.models, "models", "", "Provider models (comma-separated)")
	fs.StringVar(&options.mimeTypes, "supported-mime-types", "", "Supported MIME types (comma-separated)")
	return options
}

func addProviderBackendProbeFlags(fs *flag.FlagSet) *providerBackendCLIOptions {
	options := &providerBackendCLIOptions{}
	fs.StringVar(&options.apiKeyFile, "api-key-file", "", "Provider API key file")
	return options
}

func (options *providerBackendCLIOptions) apiKey() (string, error) {
	raw, err := os.ReadFile(options.apiKeyFile)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func (options *providerBackendCLIOptions) createOptions(fs *flag.FlagSet) ([]sdk.ProviderBackendOption, error) {
	var result []sdk.ProviderBackendOption
	var optionErr error
	fs.Visit(func(current *flag.Flag) {
		if optionErr != nil {
			return
		}
		switch current.Name {
		case "type":
			result = append(result, sdk.WithProviderBackendType(options.providerType))
		case "reasoning-control-protocol":
			result = append(result, sdk.WithProviderBackendReasoningControlProtocol(options.reasoningProtocol))
		case "api-key-file":
			value, err := options.apiKey()
			if err != nil {
				optionErr = err
				return
			}
			result = append(result, sdk.WithProviderBackendAPIKey(value))
		case "address":
			result = append(result, sdk.WithProviderBackendAddress(options.address))
		case "timeout-seconds":
			result = append(result, sdk.WithProviderBackendTimeoutSeconds(options.timeoutSeconds))
		case "models":
			result = append(result, sdk.WithProviderBackendModels(splitCSV(options.models)...))
		case "supported-mime-types":
			result = append(result, sdk.WithProviderBackendSupportedMimeTypes(splitCSV(options.mimeTypes)...))
		}
	})
	return result, optionErr
}

func (options *providerBackendCLIOptions) updateOptions(fs *flag.FlagSet) ([]sdk.ProviderBackendUpdateOption, error) {
	var result []sdk.ProviderBackendUpdateOption
	var optionErr error
	fs.Visit(func(current *flag.Flag) {
		if optionErr != nil {
			return
		}
		switch current.Name {
		case "name":
			result = append(result, sdk.WithProviderBackendUpdateName(options.name))
		case "api-key-file":
			value, err := options.apiKey()
			if err != nil {
				optionErr = err
				return
			}
			result = append(result, sdk.WithProviderBackendUpdateAPIKey(value))
		case "reasoning-control-protocol":
			result = append(result, sdk.WithProviderBackendUpdateReasoningControlProtocol(options.reasoningProtocol))
		case "timeout-seconds":
			result = append(result, sdk.WithProviderBackendUpdateTimeoutSeconds(options.timeoutSeconds))
		case "models":
			result = append(result, sdk.WithProviderBackendUpdateModels(splitCSV(options.models)...))
		case "supported-mime-types":
			result = append(result, sdk.WithProviderBackendUpdateSupportedMimeTypes(splitCSV(options.mimeTypes)...))
		}
	})
	return result, optionErr
}

type providerRouterCLIOptions struct {
	strategy                        string
	healthCheckInterval, maxRetries int32
	sessionAffinity                 bool
}

func addProviderRouterFlags(fs *flag.FlagSet) *providerRouterCLIOptions {
	options := &providerRouterCLIOptions{}
	fs.StringVar(&options.strategy, "strategy", "", "Router strategy")
	fs.Func("health-check-interval", "Router health check interval in seconds", providerBackendInt32Value(&options.healthCheckInterval))
	fs.Func("max-retries", "Router maximum retries", providerBackendInt32Value(&options.maxRetries))
	fs.BoolVar(&options.sessionAffinity, "session-affinity", false, "Enable router session affinity")
	return options
}

func (options *providerRouterCLIOptions) options(fs *flag.FlagSet) []sdk.RouterOption {
	var result []sdk.RouterOption
	fs.Visit(func(current *flag.Flag) {
		switch current.Name {
		case "strategy":
			result = append(result, sdk.WithRouterStrategy(options.strategy))
		case "health-check-interval":
			result = append(result, sdk.WithRouterHealthCheckInterval(options.healthCheckInterval))
		case "max-retries":
			result = append(result, sdk.WithRouterMaxRetries(options.maxRetries))
		case "session-affinity":
			result = append(result, sdk.WithRouterSessionAffinity(options.sessionAffinity))
		}
	})
	return result
}
