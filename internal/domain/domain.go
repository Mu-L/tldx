package domain

import (
	"fmt"

	"github.com/brandonyoungdev/tldx/internal/composer"
	"github.com/brandonyoungdev/tldx/internal/config"
	"github.com/brandonyoungdev/tldx/internal/output"
	"github.com/brandonyoungdev/tldx/internal/resolver"
)

func Exec(app *config.TldxContext, domainsOrKeywords []string) {

	composerService := composer.NewComposerService(app)
	domains, warnings := composerService.Compile(domainsOrKeywords)
	styleService := output.NewStyleService(app)
	if warnings != nil && len(warnings) > 0 {
		for _, warning := range warnings {
			if !app.Config.OnlyAvailable && app.Config.OutputFormat == "text" {
				fmt.Println(styleService.Styled(warning.Error(), "11")) // Yellow
			}
		}
	}

	resolverService := resolver.NewResolverService(app)
	resultChan := resolverService.CheckDomainsStreaming(domains)

	outputWriter := output.GetOutputWriter(app)

	output.Stat.Total = len(domains)
	for result := range resultChan {
		if result.Error != nil {
			output.Stat.Errored++
		} else if result.Available {
			output.Stat.Available++
		} else {
			output.Stat.NotAvailable++
		}
		if app.Config.OnlyAvailable && !result.Available {
			continue
		}
		outputWriter.Write(result)
	}

	outputWriter.Flush()

	if app.Config.ShowStats && app.Config.OutputFormat == "text" {
		// TODO: pipe this out for non-text formats
		fmt.Println(output.RenderStatsSummary())
	}
}
