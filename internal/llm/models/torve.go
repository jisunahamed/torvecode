package models

const ProviderTorve ModelProvider = "torve"
const ProviderTorveAnthropic ModelProvider = "torve-anthropic"

func RegisterTorveModels(catalog []Model) {
	for id, model := range SupportedModels {
		if model.Provider == ProviderTorve || model.Provider == ProviderTorveAnthropic {
			delete(SupportedModels, id)
		}
	}
	for _, model := range catalog {
		if model.Provider == "" {
			model.Provider = ProviderTorve
		}
		if model.APIModel == "" {
			model.APIModel = string(model.ID)
		}
		if model.Name == "" {
			model.Name = string(model.ID)
		}
		if model.DefaultMaxTokens <= 0 {
			model.DefaultMaxTokens = 4096
		}
		SupportedModels[model.ID] = model
	}
	ProviderPopularity[ProviderTorve] = 0
	ProviderPopularity[ProviderTorveAnthropic] = 0
}
