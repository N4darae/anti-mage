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
		s.add("permission state", "not collected")
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

	s.add("Permissions.query for notifications", valueOrAbsent(queried))
	s.add("Notification.permission", valueOrAbsent(actual))

	if queried == "" && actual == "" {
		s.Determination = Unverified
		s.conclude("neither reading was reported")
		return s
	}
	if queried == "" || actual == "" {
		s.Determination = Unverified
		s.conclude("only one of the two readings was reported")
		return s
	}
	want, known := notificationEquivalent[queried]
	if !known {
		s.Determination = Unverified
		s.conclude("the reported state is not one this project's table knows")
		return s
	}
	if want != actual {
		s.Determination = Contradiction
		s.conclude("the two interfaces report different states for one permission")
		return s
	}
	s.Determination = Consistent
	s.conclude("both interfaces report the same permission state")
	return s
}
