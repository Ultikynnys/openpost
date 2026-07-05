package commands

import (
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/openpost/cli/internal/api"
	"github.com/openpost/cli/internal/config"
)

type publicationFlags struct {
	profile  string
	title    string
	content  string
	file     string
	url      string
	accounts string
	set      string
	schedule string
	media    []string
	mediaAlt []string
	status   string
	limit    int
	offset   int
}

func newPublicationCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "publication", Short: "Create, list, validate, and publish publications"}
	cmd.AddCommand(newPublicationCreateCmd())
	cmd.AddCommand(newPublicationListCmd())
	cmd.AddCommand(newPublicationViewCmd())
	cmd.AddCommand(newPublicationValidateCmd())
	cmd.AddCommand(newPublicationPublishNowCmd())
	cmd.AddCommand(newPublicationEventsCmd())
	cmd.AddCommand(newPublicationCommentsCmd())
	return cmd
}

func newPublicationCreateCmd() *cobra.Command {
	var flags publicationFlags
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a format-first publication",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, client, workspaceID, settings, err := postRuntime(cmd)
			if err != nil {
				return err
			}
			content, err := contentFromFlags(flags.content, flags.file)
			if err != nil {
				return err
			}
			targets, err := resolveSocialTargets(cmd, client, workspaceID, flags.accounts, flags.set, true)
			if err != nil {
				return err
			}
			scheduledAt, _, err := parseScheduleFlag(cmd, client, workspaceID, targets.SetID, flags.schedule, settings.Timezone)
			if err != nil {
				return err
			}
			mediaIDs, err := resolveMedia(cmd, client, workspaceID, flags.media, flags.mediaAlt)
			if err != nil {
				return err
			}
			media := make([]api.PublicationMediaInput, 0, len(mediaIDs))
			for _, mediaID := range mediaIDs {
				media = append(media, api.PublicationMediaInput{MediaID: mediaID, Role: "attachment"})
			}
			publication, err := client.CreatePublication(cmd.Context(), api.CreatePublicationInput{
				WorkspaceID:      workspaceID,
				Title:            flags.title,
				ContentProfile:   defaultString(flags.profile, "short_text"),
				SourceText:       content,
				SourceURL:        flags.url,
				ScheduledAt:      scheduledAt,
				SocialAccountIDs: targets.AccountIDs,
				Media:            media,
				Metadata:         map[string]interface{}{"created_from": "cli"},
			})
			if err != nil {
				return err
			}
			return printPublicationSummary(cfg, publication)
		},
	}
	cmd.Flags().StringVar(&flags.profile, "profile", "short_text", "content profile: short_text, thread, link_share, image_post, carousel, story, short_video, long_video")
	cmd.Flags().StringVar(&flags.title, "title", "", "publication title")
	cmd.Flags().StringVar(&flags.content, "content", "", "source text")
	cmd.Flags().StringVar(&flags.file, "file", "", "read source text from file or '-' for stdin")
	cmd.Flags().StringVar(&flags.url, "url", "", "source URL for link shares")
	cmd.Flags().StringVar(&flags.accounts, "accounts", "", "comma-separated account IDs/slugs/platforms")
	cmd.Flags().StringVar(&flags.set, "set", "", "social media set name or ID")
	cmd.Flags().StringVar(&flags.schedule, "schedule", "", "schedule time")
	cmd.Flags().StringArrayVar(&flags.media, "media", nil, "media ID/path/URL to attach")
	cmd.Flags().StringArrayVar(&flags.mediaAlt, "media-alt", nil, "alt text for uploaded media")
	return cmd
}

func newPublicationListCmd() *cobra.Command {
	var flags publicationFlags
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List publications",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, client, workspaceID, _, err := postRuntime(cmd)
			if err != nil {
				return err
			}
			publications, err := client.ListPublications(cmd.Context(), api.ListPublicationsInput{
				WorkspaceID:    workspaceID,
				Status:         flags.status,
				ContentProfile: flags.profile,
				Limit:          flags.limit,
				Offset:         flags.offset,
			})
			if err != nil {
				return err
			}
			p := printerFrom(cfg)
			if cfg.AsJSON {
				return p.PrintJSON(publications)
			}
			rows := make([][]string, 0, len(publications))
			for _, publication := range publications {
				rows = append(rows, []string{
					publication.ID,
					publication.Status,
					publication.ContentProfile,
					scheduleLabel(publication.ScheduledAt),
					preview(publication.Title, 40),
					strconv.Itoa(len(publication.Renditions)),
				})
			}
			p.Table([]string{"ID", "STATUS", "PROFILE", "SCHEDULED", "TITLE", "RENDITIONS"}, rows)
			return nil
		},
	}
	cmd.Flags().StringVar(&flags.status, "status", "", "filter by status")
	cmd.Flags().StringVar(&flags.profile, "profile", "", "filter by content profile")
	cmd.Flags().IntVar(&flags.limit, "limit", 0, "maximum number of publications to return")
	cmd.Flags().IntVar(&flags.offset, "offset", 0, "number of publications to skip")
	return cmd
}

func newPublicationViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "view <publication-id>",
		Short: "View a publication",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := runtimeFrom(cmd)
			if err != nil {
				return err
			}
			client, err := clientFrom(cfg)
			if err != nil {
				return err
			}
			publication, err := client.GetPublication(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return printPublicationSummary(cfg, publication)
		},
	}
}

func newPublicationValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate <publication-id>",
		Short: "Validate a publication",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := runtimeFrom(cmd)
			if err != nil {
				return err
			}
			client, err := clientFrom(cfg)
			if err != nil {
				return err
			}
			result, err := client.ValidatePublication(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			p := printerFrom(cfg)
			if cfg.AsJSON {
				return p.PrintJSON(result)
			}
			p.Printf("valid\t%t", result.Valid)
			for _, issue := range result.Issues {
				p.Printf("%s\t%s\t%s", issue.Severity, issue.Code, issue.Message)
			}
			return nil
		},
	}
}

func newPublicationPublishNowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "publish-now <publication-id>",
		Short: "Queue a publication for immediate publishing",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := runtimeFrom(cmd)
			if err != nil {
				return err
			}
			client, err := clientFrom(cfg)
			if err != nil {
				return err
			}
			result, err := client.PublishPublicationNow(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			p := printerFrom(cfg)
			if cfg.AsJSON {
				return p.PrintJSON(result)
			}
			p.Printf("%s\t%s", result.Message, emptyDash(result.JobID))
			return nil
		},
	}
}

func newPublicationEventsCmd() *cobra.Command {
	var flags publicationFlags
	cmd := &cobra.Command{
		Use:   "events <publication-id>",
		Short: "List publication lifecycle events",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := runtimeFrom(cmd)
			if err != nil {
				return err
			}
			client, err := clientFrom(cfg)
			if err != nil {
				return err
			}
			events, err := client.ListPublicationEvents(cmd.Context(), args[0], flags.limit)
			if err != nil {
				return err
			}
			p := printerFrom(cfg)
			if cfg.AsJSON {
				return p.PrintJSON(events)
			}
			rows := make([][]string, 0, len(events))
			for _, event := range events {
				rows = append(rows, []string{
					event.CreatedAt,
					event.Type,
					event.Status,
					emptyDash(event.RenditionID),
					event.Message,
				})
			}
			p.Table([]string{"CREATED", "TYPE", "STATUS", "RENDITION", "MESSAGE"}, rows)
			return nil
		},
	}
	cmd.Flags().IntVar(&flags.limit, "limit", 0, "maximum number of events to return")
	return cmd
}

func newPublicationCommentsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "comments <rendition-id>",
		Short: "List comments for a published rendition",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := runtimeFrom(cmd)
			if err != nil {
				return err
			}
			client, err := clientFrom(cfg)
			if err != nil {
				return err
			}
			comments, err := client.ListRenditionComments(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			p := printerFrom(cfg)
			if cfg.AsJSON {
				return p.PrintJSON(comments)
			}
			rows := make([][]string, 0, len(comments))
			for _, comment := range comments {
				rows = append(rows, []string{
					comment.ID,
					emptyDash(comment.AuthorName),
					hiddenLabel(comment.Hidden),
					commentActions(comment),
					preview(comment.Text, 80),
				})
			}
			p.Table([]string{"ID", "AUTHOR", "HIDDEN", "ACTIONS", "TEXT"}, rows)
			return nil
		},
	}
}

func printPublicationSummary(cfg *config.Runtime, publication *api.Publication) error {
	p := printerFrom(cfg)
	if cfg.AsJSON {
		return p.PrintJSON(publication)
	}
	p.Table([]string{"FIELD", "VALUE"}, [][]string{
		{"id", publication.ID},
		{"workspace_id", publication.WorkspaceID},
		{"status", publication.Status},
		{"profile", publication.ContentProfile},
		{"scheduled_at", scheduleLabel(publication.ScheduledAt)},
		{"title", publication.Title},
		{"rendition_count", strconv.Itoa(len(publication.Renditions))},
	})
	for _, rendition := range publication.Renditions {
		p.Printf("rendition %s\t%s\t%s\t%s", rendition.ID, rendition.Platform, rendition.Status, emptyDash(rendition.ErrorMessage))
	}
	return nil
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func hiddenLabel(hidden bool) string {
	if hidden {
		return "yes"
	}
	return "no"
}

func commentActions(comment api.Comment) string {
	actions := make([]string, 0, 3)
	if comment.CanReply {
		actions = append(actions, "reply")
	}
	if comment.CanHide {
		actions = append(actions, "hide")
	}
	if comment.CanDelete {
		actions = append(actions, "delete")
	}
	if len(actions) == 0 {
		return "-"
	}
	return strings.Join(actions, ",")
}
