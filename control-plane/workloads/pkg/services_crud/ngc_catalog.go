package services_crud

import (
	"sort"
	"strings"

	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
)

// NimModelInfo holds information about a NIM model
type NimModelInfo struct {
	ID          string
	Description string
}

// Hardcoded NIM models list with real descriptions as fallback when NGC API fails
var HardcodedNimModels = []NimModelInfo{
	{ID: "hive/ai-generated-image-detection", Description: "Detects AI-generated images with high accuracy"},
	{ID: "deepmind/alphafold2", Description: "Predicts protein structures from amino acid sequences"},
	{ID: "deepmind/alphafold2-multimer", Description: "Predicts protein complex structures from amino acid sequences"},
	{ID: "nvidia/parakeet-1-1b-ctc-en-us", Description: "English speech recognition model using CTC decoding"},
	{ID: "nvidia/parakeet-1-1b-rnnt-multilingual", Description: "Multilingual speech recognition model using RNN-T"},
	{ID: "nvidia/parakeet-ctc-1.1b-asr", Description: "Automatic speech recognition model with CTC decoding"},
	{ID: "nvidia/maxine-audio2face-2d", Description: "Converts audio to 2D facial expressions and lip sync"},
	{ID: "nvidia/audio2face-3d", Description: "Converts audio to 3D facial expressions and lip sync"},
	{ID: "baai/bge-large-zh-v1-5", Description: "Large-scale Chinese text embedding model"},
	{ID: "baai/bge-m3", Description: "Multilingual text embedding model supporting 100+ languages"},
	{ID: "meta/codellama-13b-instruct", Description: "13B parameter code generation model with instruction tuning"},
	{ID: "meta/codellama-34b-instruct", Description: "34B parameter code generation model with instruction tuning"},
	{ID: "meta/codellama-70b-instruct", Description: "70B parameter code generation model with instruction tuning"},
	{ID: "nvidia/cosmos-predict1-7b-text2world", Description: "Generates physics-aware world states from text prompts"},
	{ID: "nvidia/cosmos-predict1-7b-video2world", Description: "Generates physics-aware world states from video inputs"},
	{ID: "hive/deepfake-image-detection", Description: "Detects deepfake and manipulated images"},
	{ID: "deepseek-ai/deepseek-r1", Description: "Reasoning-focused language model for complex problem solving"},
	{ID: "deepseek-ai/deepseek-r1-distill-llama-70b", Description: "Distilled reasoning model based on Llama-70B"},
	{ID: "deepseek-ai/deepseek-r1-distill-llama-8b", Description: "Distilled reasoning model based on Llama-8B"},
	{ID: "deepseek-ai/deepseek-r1-distill-qwen-14b", Description: "Distilled reasoning model based on Qwen-14B"},
	{ID: "deepseek-ai/deepseek-r1-distill-qwen-32b", Description: "Distilled reasoning model based on Qwen-32B"},
	{ID: "deepseek-ai/deepseek-r1-distill-qwen-7b", Description: "Distilled reasoning model based on Qwen-7B"},
	{ID: "mit/diffdock", Description: "Protein-ligand docking using diffusion models"},
	{ID: "nvidia/domino-automotive-aero", Description: "Automotive aerodynamic simulation and optimization"},
	{ID: "nvidia/corrdiff", Description: "Correlation-based diffusion model for time series forecasting"},
	{ID: "nvidia/fourcastnet", Description: "Global weather forecasting using neural networks"},
	{ID: "arc/evo2-40b", Description: "40B parameter language model optimized for reasoning"},
	{ID: "nvidia/maxine-eye-contact", Description: "Real-time eye contact correction for video calls"},
	{ID: "black-forest-labs/flux.1-dev", Description: "Advanced reasoning and problem-solving model"},
	{ID: "google/gemma-2-2b-instruct", Description: "2B parameter instruction-tuned language model"},
	{ID: "google/gemma-2-9b-it", Description: "9B parameter instruction-tuned language model"},
	{ID: "nvidia/llama-3.1-nemoguard-8b-content-safety", Description: "Content safety model based on Llama-3.1"},
	{ID: "nvidia/llama-3.1-nemoguard-8b-topic-control", Description: "Topic control model based on Llama-3.1"},
	{ID: "defog/llama-3-sqlcoder-8b", Description: "SQL code generation model based on Llama-3"},
	{ID: "tokyotech-llm/llama-3-swallow-70b-instruct-v0.1", Description: "70B Japanese language model based on Llama-3"},
	{ID: "yentinglin/llama-3-taiwan-70b-instruct", Description: "70B Traditional Chinese language model based on Llama-3"},
	{ID: "meta/llama-3.1-405b-instruct", Description: "405B parameter instruction-tuned language model"},
	{ID: "meta/llama-3.1-70b-instruct", Description: "70B parameter instruction-tuned language model"},
	{ID: "meta/llama-3.1-70b-instruct-pb24h2", Description: "70B parameter instruction-tuned model with 24-hour pre-training"},
	{ID: "meta/llama-3.1-8b-base", Description: "8B parameter base language model"},
	{ID: "meta/llama-3.1-8b-instruct", Description: "8B parameter instruction-tuned language model"},
	{ID: "meta/llama-3.1-8b-instruct-pb24h2", Description: "8B parameter instruction-tuned model with 24-hour pre-training"},
	{ID: "nvidia/llama-3.1-nemotron-70b-instruct", Description: "70B instruction-tuned model for coding and reasoning"},
	{ID: "nvidia/llama-3.1-nemotron-nano-8b-v1", Description: "8B parameter efficient instruction-tuned model"},
	{ID: "tokyotech-llm/llama-3.1-swallow-70b-instruct-v0.1", Description: "70B Japanese instruction-tuned model based on Llama-3.1"},
	{ID: "tokyotech-llm/llama-3.1-swallow-8b-instruct-v0.1", Description: "8B Japanese instruction-tuned model based on Llama-3.1"},
	{ID: "meta/llama-3.2-11b-vision-instruct", Description: "11B parameter vision-language instruction model"},
	{ID: "meta/llama-3.2-1b-instruct", Description: "1B parameter instruction-tuned language model"},
	{ID: "meta/llama-3.2-3b-instruct", Description: "3B parameter instruction-tuned language model"},
	{ID: "meta/llama-3.2-90b-vision-instruct", Description: "90B parameter vision-language instruction model"},
	{ID: "meta/llama-3.3-70b-instruct", Description: "70B parameter instruction-tuned language model"},
	{ID: "nvidia/llama-3.3-nemotron-super-49b-v1", Description: "49B parameter super-efficient instruction-tuned model"},
	{ID: "nvidia/magpie-tts-multilingual", Description: "Multilingual text-to-speech synthesis model"},
	{ID: "nvidia/maisi", Description: "Multimodal AI system for image and text understanding"},
	{ID: "meta/llama-2-13b-chat", Description: "13B parameter chat-optimized language model"},
	{ID: "meta/llama-2-70b-chat", Description: "70B parameter chat-optimized language model"},
	{ID: "meta/llama-2-7b-chat", Description: "7B parameter chat-optimized language model"},
	{ID: "meta/llama3-70b-instruct", Description: "70B parameter instruction-tuned language model"},
	{ID: "meta/llama3-8b-instruct", Description: "8B parameter instruction-tuned language model"},
	{ID: "mistralai/mistral-7b-instruct-v0.3", Description: "7B parameter instruction-tuned language model"},
	{ID: "nv-mistralai/mistral-nemo-12b-instruct", Description: "12B parameter instruction-tuned model based on Mistral"},
	{ID: "nv-mistralai/mistral-nemo-minitron-8b-8k-instruct", Description: "8B parameter model with 8K context window"},
	{ID: "mistralai/mixtral-8x22b-instruct-v01", Description: "Mixture of experts model with 22B parameters"},
	{ID: "mistralai/mixtral-8x7b-instruct-v01", Description: "Mixture of experts model with 7B parameters"},
	{ID: "nvidia/molmim", Description: "Molecular modeling and simulation using machine learning"},
	{ID: "colabfold/msa-search", Description: "Multiple sequence alignment search for protein structure prediction"},
	{ID: "nvidia/nemoretriever-graphic-elements-v1", Description: "Retrieval model for graphic elements in documents"},
	{ID: "baidu/paddleocr", Description: "Optical character recognition for multiple languages"},
	{ID: "nvidia/nemoretriever-page-elements-v2", Description: "Retrieval model for page elements in documents"},
	{ID: "nvidia/nemoretriever-table-structure-v1", Description: "Table structure recognition and understanding"},
	{ID: "nvidia/nv-yolox-page-elements-v1", Description: "Page element detection using YOLOX architecture"},
	{ID: "nvidia/nemoguard-jailbreak-detect", Description: "Jailbreak detection for language model safety"},
	{ID: "nvidia/nemoretriever-parse", Description: "Document parsing and information extraction"},
	{ID: "nvidia/nemotron-4-340b-instruct", Description: "340B parameter instruction-tuned language model"},
	{ID: "nvidia/nemotron-4-340b-reward", Description: "340B parameter reward model for reinforcement learning"},
	{ID: "nvidia/megatron-1b-nmt", Description: "1B parameter neural machine translation model"},
	{ID: "nvidia/nvclip", Description: "Contrastive language-image pre-training model"},
	{ID: "nvidia/genmol", Description: "Generative molecular design and optimization"},
	{ID: "nvidia/nv-embedqa-e5-v5", Description: "Question-answering embedding model"},
	{ID: "nvidia/nv-embedqa-e5-v5-pb24h2", Description: "Question-answering embedding model with 24-hour pre-training"},
	{ID: "nvidia/llama-3.2-nv-embedqa-1b-v2", Description: "1B parameter question-answering model based on Llama-3.2"},
	{ID: "nvidia/llama-3.2-nv-rerankqa-1b-v2", Description: "1B parameter question-answering reranking model"},
	{ID: "nvidia/nv-rerankqa-mistral-4b-v3", Description: "4B parameter question-answering reranking model based on Mistral"},
	{ID: "nvidia/nv-embedqa-mistral-7b-v2", Description: "7B parameter question-answering embedding model based on Mistral"},
	{ID: "openfold/openfold2", Description: "Protein structure prediction using deep learning"},
	{ID: "nvidia/parakeet-0-6b-ctc-en-us", Description: "6B parameter English speech recognition model"},
	{ID: "microsoft/phi-3-mini-4k-instruct", Description: "Mini instruction-tuned language model with 4K context"},
	{ID: "phind/phind-codellama-34b-v2-instruct", Description: "34B parameter code generation model"},
	{ID: "ipd/proteinmpnn", Description: "Protein structure design using message passing neural networks"},
	{ID: "qwen/qwen-2.5-72b-instruct", Description: "72B parameter instruction-tuned language model"},
	{ID: "qwen/qwen-2.5-7b-instruct", Description: "7B parameter instruction-tuned language model"},
	{ID: "ipd/rfdiffusion", Description: "Protein structure generation using diffusion models"},
	{ID: "nvidia/riva-asr", Description: "Automatic speech recognition for real-time applications"},
	{ID: "nvidia/riva-nmt", Description: "Neural machine translation for real-time applications"},
	{ID: "nvidia/riva-tts", Description: "Text-to-speech synthesis for real-time applications"},
	{ID: "snowflake/arctic-embed-l", Description: "Large-scale text embedding model for enterprise applications"},
	{ID: "bigcode/starcoder2-7b", Description: "7B parameter code generation model"},
	{ID: "bigcode/starcoderbase-15b", Description: "15B parameter base code generation model"},
	{ID: "nvidia/maxine-studio-voice", Description: "Professional voice enhancement and processing"},
	{ID: "nvidia/fastpitch-hifigan-tts", Description: "Fast text-to-speech synthesis using FastPitch and HiFi-GAN"},
}

// NGC API query struct for the 'q' parameter
// Only the fields we need
// See: https://api.ngc.nvidia.com/v2/search/catalog/resources/CONTAINER?q=...
type ngcQuery struct {
	Filters    []ngcFilter  `json:"filters"`
	OrderBy    []ngcOrderBy `json:"orderBy"`
	Page       int          `json:"page"`
	PageSize   int          `json:"pageSize"`
	Query      string       `json:"query"`
	ScoredSize int          `json:"scoredSize"`
}

type ngcFilter struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

type ngcOrderBy struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

// getFallbackNimCatalog returns a hardcoded container catalog when the NGC API fails
func getFallbackNimCatalog() api.Catalog {
	var resources []api.CatalogResourcesInner

	for _, model := range HardcodedNimModels {
		// Parse org and name from modelId (format: "org/name")
		parts := strings.Split(model.ID, "/")
		orgName := ""
		name := model.ID
		if len(parts) == 2 {
			orgName = parts[0]
			name = parts[1]
		}

		// Apply the same validation as in parseNgcResponse
		validatedResourceId := model.ID
		if model.ID != "" {
			parts := strings.Split(model.ID, "/")
			// If resource ID has 3 parts and the first part equals the orgName, remove the first part
			if len(parts) == 3 && parts[0] == orgName {
				validatedResourceId = parts[1] + "/" + parts[2]
			}
		}

		resource := api.CatalogResourcesInner{
			OrgName:      orgName,
			ResourceId:   validatedResourceId,
			ResourceType: "CONTAINER",
			DisplayName:  name,
			Name:         name,
			Description:  model.Description,
			IsPublic:     false,
		}
		resources = append(resources, resource)
	}

	// Sort resources alphabetically by resource ID
	sort.Slice(resources, func(i, j int) bool {
		return resources[i].ResourceId < resources[j].ResourceId
	})

	return api.Catalog{
		IsStaticData:    true,
		ResultTotal:     int32(len(resources)),
		ResultPageTotal: int32(len(resources)),
		Resources:       resources,
	}
}
