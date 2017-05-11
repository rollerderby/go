package ruleset

func (rs *Ruleset) NewRule(Name, Type, DefaultValue string, EnumValues []string) (*Rule, error) {
	r, err := rs.Rules().New(Name)
	if err != nil {
		return nil, err
	}
	if err = r.SetType(Type); err != nil {
		return r, err
	}
	if err = r.SetDefaultValue(DefaultValue); err != nil {
		return r, err
	}
	if err = r.SetValue(DefaultValue); err != nil {
		return r, err
	}
	for _, enum := range EnumValues {
		if err := r.EnumValues().Add(enum); err != nil {
			return r, err
		}
	}

	return r, nil
}

func (h *Ruleset) init() error {
	h.NewRule("Clock.Intermission.Direction", "Enum", "Count Down", []string{"Count Up", "Count Down"})
	h.NewRule("Clock.Intermission.MaximumNumber", "Integer", "2", nil)
	h.NewRule("Clock.Intermission.MaximumTime", "Time", "60:00", nil)
	h.NewRule("Clock.Intermission.MinimumNumber", "Integer", "0", nil)
	h.NewRule("Clock.Intermission.MinimumTime", "Time", "0:00", nil)
	h.NewRule("Clock.Intermission.Name", "String", "Intermission", nil)
	h.NewRule("Clock.Intermission.Time", "Time", "15:00", nil)
	h.NewRule("Clock.Jam.Direction", "Enum", "Count Down", []string{"Count Up", "Count Down"})
	h.NewRule("Clock.Jam.MaximumNumber", "Integer", "999", nil)
	h.NewRule("Clock.Jam.MaximumTime", "Time", "2:00", nil)
	h.NewRule("Clock.Jam.MinimumNumber", "Integer", "1", nil)
	h.NewRule("Clock.Jam.MinimumTime", "Time", "0:00", nil)
	h.NewRule("Clock.Jam.Name", "String", "Jam", nil)
	h.NewRule("Clock.Lineup.Direction", "Enum", "Count Up", []string{"Count Up", "Count Down"})
	h.NewRule("Clock.Lineup.MaximumNumber", "Integer", "999", nil)
	h.NewRule("Clock.Lineup.MaximumTime", "Time", "60:00", nil)
	h.NewRule("Clock.Lineup.MinimumNumber", "Integer", "1", nil)
	h.NewRule("Clock.Lineup.MinimumTime", "Time", "0:00", nil)
	h.NewRule("Clock.Lineup.Name", "String", "Lineup", nil)
	h.NewRule("Clock.Period.Direction", "Enum", "Count Down", []string{"Count Up", "Count Down"})
	h.NewRule("Clock.Period.MaximumNumber", "Integer", "2", nil)
	h.NewRule("Clock.Period.MaximumTime", "Time", "30:00", nil)
	h.NewRule("Clock.Period.MinimumNumber", "Integer", "1", nil)
	h.NewRule("Clock.Period.Name", "String", "Period", nil)
	h.NewRule("Clock.Timeout.Direction", "Enum", "Count Up", []string{"Count Up", "Count Down"})
	h.NewRule("Clock.Timeout.MaximumNumber", "Integer", "999", nil)
	h.NewRule("Clock.Timeout.MaximumTime", "Time", "60:00", nil)
	h.NewRule("Clock.Timeout.MinimumNumber", "Integer", "1", nil)
	h.NewRule("Clock.Timeout.MinimumTime", "Time", "0:00", nil)
	h.NewRule("Clock.Timeout.Name", "String", "Timeout", nil)
	h.NewRule("ScoreBoard.BackgroundStyle", "String", "", nil)
	h.NewRule("ScoreBoard.BoxStyle", "String", "box_flat", nil)
	h.NewRule("ScoreBoard.CurrentView", "String", "scoreboard", nil)
	h.NewRule("ScoreBoard.CustomHtml", "String", "/customhtml/fullscreen/example.html", nil)
	h.NewRule("ScoreBoard.HideJamTotals", "Boolean", "Show Jam Totals", nil)
	h.NewRule("ScoreBoard.Image", "String", "/images/fullscreen/American Flag.jpg", nil)
	h.NewRule("ScoreBoard.Intermission.Intermission", "String", "Intermission", nil)
	h.NewRule("ScoreBoard.Intermission.Official", "String", "Final Score", nil)
	h.NewRule("ScoreBoard.Intermission.PreGame", "String", "Time To Derby", nil)
	h.NewRule("ScoreBoard.Intermission.Unofficial", "String", "Unofficial Score", nil)
	h.NewRule("ScoreBoard.SidePadding", "Integer", "0", nil)
	h.NewRule("ScoreBoard.SwapTeams", "Boolean", "Teams Normal", nil)
	h.NewRule("ScoreBoard.Video", "String", "/videos/fullscreen/American Flag.webm", nil)
	h.NewRule("Team.1.Name", "String", "Team 1", nil)
	h.NewRule("Team.2.Name", "String", "Team 2", nil)
	h.NewRule("Team.OfficialReviews", "Integer", "1", nil)
	h.NewRule("Team.Timeouts", "Integer", "3", nil)

	return nil
}
