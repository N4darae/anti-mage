package scan

var notificationEquivalent = map[string]string{
	"granted": "granted",
	"denied":  "denied",
	"prompt":  "default",
}

func sectionPermissions(r Request, _ Inputs, _ claim) Section {
	s := Section{Determination: Inconclusive}

	raw, ok := r.value("perm.state")
	if !ok {
		s.Determination = Unverified
		s.Rows = append(s.Rows, Row{
			Label: "permission state",
			Value: "not collected",
			Note:  "this reading compares two interfaces reporting the same permission; a payload that names neither leaves it nothing to apply to, so it carries no weight either way",
		})
		return s
	}

	queried := firstString(raw,
		[]string{"notifications", "query"},
		[]string{"notifications", "state"},
		[]string{"notifications"},
		[]string{"queryNotifications"},
		[]string{"permQueryNotifications"},
	)
	actual := firstString(raw,
		[]string{"notifications", "actual"},
		[]string{"notificationPermission"},
		[]string{"notificationsActual"},
	)

	s.Rows = append(s.Rows, Row{Label: "Permissions.query for notifications", Value: valueOrAbsent(queried), Note: "the permission state"})
	s.Rows = append(s.Rows, Row{Label: "Notification.permission", Value: valueOrAbsent(actual), Note: "the same permission, read through the other interface"})

	if queried == "" && actual == "" {
		s.Determination = Unverified
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "neither reading was reported",
			Note:  "an observation was collected but named neither interface, so this reading has nothing to apply to",
		})
		return s
	}
	if queried == "" || actual == "" {
		s.Determination = Unverified
		s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "only one of the two readings was reported", Note: "a comparison needs both, so this reading has nothing to apply to"})
		return s
	}
	want, known := notificationEquivalent[queried]
	if !known {
		s.Determination = Unverified
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "the reported state is not one this project's table knows",
			Note:  "a state this project cannot place carries no weight, in either direction",
		})
		return s
	}
	if want != actual {
		s.Determination = Contradiction
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "the two interfaces report different states for one permission",
			Note:  "the permission state " + clip(queried, 40) + " corresponds to " + want + ", not " + clip(actual, 40),
		})
		return s
	}
	s.Determination = Consistent
	s.Rows = append(s.Rows, Row{
		Label: "conclusion",
		Value: "both interfaces report the same permission state",
		Note:  "",
	})
	return s
}

func firstString(v any, paths ...[]string) string {
	for _, p := range paths {
		if s, ok := str(v, p...); ok && s != "" {
			return s
		}
	}
	return ""
}

func valueOrAbsent(s string) string {
	if s == "" {
		return "not reported"
	}
	return clip(s, 40)
}
