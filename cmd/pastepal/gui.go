package main

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/JacobRWebb/PastePal-OS/internal/core"
	"github.com/JacobRWebb/PastePal-OS/internal/models"
)

// GUI represents the graphical user interface for PastePal
type GUI struct {
	app        fyne.App
	mainWindow fyne.Window
	pasteApp   *core.PastePalApp

	// Current UI state
	currentContainer fyne.CanvasObject
}

// NewGUI creates a new GUI instance
func NewGUI(pasteApp *core.PastePalApp) *GUI {
	// Use app.NewWithID instead of app.New to provide a unique ID for preferences
	fyneApp := app.NewWithID("com.pastepal.app")
	fyneApp.Settings().SetTheme(theme.DarkTheme())

	mainWindow := fyneApp.NewWindow("PastePal - Zero Knowledge Security")
	mainWindow.Resize(fyne.NewSize(800, 600))

	return &GUI{
		app:        fyneApp,
		mainWindow: mainWindow,
		pasteApp:   pasteApp,
	}
}

// Run starts the GUI application
func (g *GUI) Run() {
	// Set initial screen based on login state
	if g.pasteApp.IsLoggedIn {
		g.showDashboard()
	} else {
		g.showLoginScreen()
	}

	g.mainWindow.ShowAndRun()
}

// showLoginScreen displays the login screen
func (g *GUI) showLoginScreen() {
	// Create input fields
	emailEntry := widget.NewEntry()
	emailEntry.SetPlaceHolder("Email")
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")

	// Remember me checkbox
	rememberMeCheck := widget.NewCheck("Remember Me", nil)

	// Status label for showing login progress within the app
	statusLabel := widget.NewLabelWithStyle(
		"",
		fyne.TextAlignCenter,
		fyne.TextStyle{},
	)
	statusLabel.Hide()

	// Create login button outside of its definition to avoid scope issues
	var loginButton *widget.Button
	loginButton = widget.NewButton("Login", func() {
		email := strings.TrimSpace(emailEntry.Text)
		password := strings.TrimSpace(passwordEntry.Text)

		if email == "" || password == "" {
			dialog.ShowError(fmt.Errorf("email and password are required"), g.mainWindow)
			return
		}

		// Show status in the app instead of notification
		statusLabel.SetText("Logging in...")
		statusLabel.Show()
		loginButton.Disable()

		go func() {
			err := g.pasteApp.Login(email, password, rememberMeCheck.Checked)
			
			// Update UI in the main thread
			g.mainWindow.Canvas().Refresh(g.currentContainer)
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("Login failed: %v", err))
				loginButton.Enable()
				return
			}
			statusLabel.Hide()
			loginButton.Enable()
			g.showDashboard()
		}()
	})

	registerBtn := widget.NewButton("Register", func() {
		g.showRegisterScreen()
	})

	// Try to load saved credentials
	go func() {
		email, _, err := g.pasteApp.LocalStorage.GetSavedCredentials()
		if err == nil && email != "" {
			emailEntry.SetText(email)
			rememberMeCheck.SetChecked(true)
			// Focus on password field since email is already filled
			passwordEntry.FocusGained()
		}
	}()

	// Create form layout
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Email", Widget: emailEntry},
			{Text: "Password", Widget: passwordEntry},
			{Text: "", Widget: rememberMeCheck},
		},
		OnSubmit: loginButton.OnTapped,
	}

	// Create a container with all the elements
	content := container.NewVBox(
		container.NewCenter(widget.NewLabelWithStyle("PastePal", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		container.NewCenter(widget.NewLabelWithStyle("Secure Zero-Knowledge Paste Sharing", fyne.TextAlignCenter, fyne.TextStyle{Italic: true})),
		form,
		container.NewHBox(layout.NewSpacer(), loginButton, registerBtn, layout.NewSpacer()),
		statusLabel,
	)

	// Try auto-login if credentials are saved
	go func() {
		if g.pasteApp.AutoLogin() {
			// Successfully auto-logged in, show dashboard
			g.showDashboard()
		}
	}()

	g.currentContainer = content
	g.mainWindow.SetContent(content)
}

// showRegisterScreen displays the registration screen
func (g *GUI) showRegisterScreen() {
	title := widget.NewLabelWithStyle(
		"Create Account",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)
	titleStyled := container.NewCenter(container.NewPadded(title))

	emailEntry := widget.NewEntry()
	emailEntry.SetPlaceHolder("Email")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")

	confirmPasswordEntry := widget.NewPasswordEntry()
	confirmPasswordEntry.SetPlaceHolder("Confirm Password")

	// Status label for showing registration progress
	statusLabel := widget.NewLabelWithStyle(
		"",
		fyne.TextAlignCenter,
		fyne.TextStyle{},
	)
	statusLabel.Hide()

	// Create register button outside of its definition to avoid scope issues
	var registerButton *widget.Button
	registerButton = widget.NewButton("Register", func() {
		email := strings.TrimSpace(emailEntry.Text)
		password := strings.TrimSpace(passwordEntry.Text)
		confirmPassword := strings.TrimSpace(confirmPasswordEntry.Text)

		if email == "" || password == "" {
			dialog.ShowError(fmt.Errorf("email and password are required"), g.mainWindow)
			return
		}

		if password != confirmPassword {
			dialog.ShowError(fmt.Errorf("passwords do not match"), g.mainWindow)
			return
		}

		// Show status in the app instead of progress dialog
		statusLabel.SetText("Creating your account...")
		statusLabel.Show()
		registerButton.Disable()

		go func() {
			err := g.pasteApp.Register(email, password)
			// Update UI in a goroutine-safe way
			g.mainWindow.Canvas().Refresh(g.currentContainer)
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("Registration failed: %v", err))
				registerButton.Enable()
				return
			}

			statusLabel.Hide()
			registerButton.Enable()
			dialog.ShowInformation("Registration Successful", "Your account has been created. You can now log in.", g.mainWindow)
			g.showLoginScreen()
		}()
	})

	backButton := widget.NewButton("Back to Login", func() {
		g.showLoginScreen()
	})

	// Create a more professional card-like container
	registerCard := container.NewVBox(
		widget.NewLabelWithStyle("Create a new account", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		container.NewPadded(emailEntry),
		container.NewPadded(passwordEntry),
		container.NewPadded(confirmPasswordEntry),
		statusLabel,
		container.NewPadded(registerButton),
		container.NewHBox(
			layout.NewSpacer(),
			backButton,
			layout.NewSpacer(),
		),
	)

	// Add padding and styling to make it look like a card
	registerCardStyled := container.NewPadded(
		container.NewVBox(
			layout.NewSpacer(),
			titleStyled,
			container.NewPadded(registerCard),
			layout.NewSpacer(),
		),
	)

	g.currentContainer = registerCardStyled
	g.mainWindow.SetContent(registerCardStyled)
}

// showDashboard displays the main dashboard after login
func (g *GUI) showDashboard() {
	header := g.createHeader()

	// Create tabs for different sections with improved styling
	tabs := container.NewAppTabs(
		container.NewTabItem("My Pastes", g.createPastesListTab()),
		container.NewTabItem("Create Paste", g.createNewPasteTab()),
		container.NewTabItem("Account", g.createAccountTab()),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	content := container.NewBorder(header, nil, nil, nil, tabs)
	g.currentContainer = content
	g.mainWindow.SetContent(content)
}

// createHeader creates the header with app title and logout button
func (g *GUI) createHeader() fyne.CanvasObject {
	title := widget.NewLabelWithStyle(
		"PastePal",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)
	titleStyled := container.NewCenter(container.NewPadded(title))

	// Create a more professional logout button with icon
	logoutBtn := widget.NewButtonWithIcon("Logout", theme.LogoutIcon(), func() {
		dialog.ShowConfirm(
			"Confirm Logout",
			"Are you sure you want to log out?",
			func(confirm bool) {
				if confirm {
					g.pasteApp.Logout()
					g.showLoginScreen()
				}
			},
			g.mainWindow,
		)
	})

	// Create a more professional header with subtle separator
	header := container.NewBorder(
		nil,
		container.NewHBox(
			container.NewHBox(layout.NewSpacer()),
			widget.NewSeparator(),
		),
		logoutBtn,
		nil,
		titleStyled,
	)

	return header
}

// createPastesListTab creates the tab for listing user's pastes
func (g *GUI) createPastesListTab() fyne.CanvasObject {
	// Create a more professional refresh button with icon and tooltip
	refreshBtn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), nil)
	refreshBtn.Importance = widget.MediumImportance
	
	// Create a styled heading for the tab
	heading := widget.NewLabelWithStyle(
		"Your Pastes", 
		fyne.TextAlignLeading, 
		fyne.TextStyle{Bold: true},
	)

	// Create a more professional list with better styling
	list := widget.NewList(
		func() int { return 0 }, // Will be updated in refreshPastesList
		func() fyne.CanvasObject {
			// Create a more professional list item template
			titleLabel := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			dateLabel := widget.NewLabelWithStyle("", fyne.TextAlignTrailing, fyne.TextStyle{})
			
			return container.NewBorder(
				nil, nil, nil, dateLabel,
				titleLabel,
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			// Will be updated in refreshPastesList
		},
	)

	refreshBtn.OnTapped = func() {
		g.refreshPastesList(list)
	}

	// Initial load of pastes
	g.refreshPastesList(list)

	// Create a more professional empty state message
	noContentLabel := widget.NewLabelWithStyle(
		"No pastes found. Create your first paste!",
		fyne.TextAlignCenter,
		fyne.TextStyle{Italic: true},
	)
	noContentLabel.Hide()

	// Create a more professional header for the tab
	header := container.NewBorder(
		nil, nil, heading, refreshBtn,
		widget.NewSeparator(),
	)

	// Create a more professional container with proper spacing
	container := container.NewBorder(
		header,
		nil, nil, nil,
		container.NewStack(list, noContentLabel),
	)

	return container
}

// refreshPastesList refreshes the list of pastes
func (g *GUI) refreshPastesList(list *widget.List) {
	// Create a status label instead of a progress dialog
	statusLabel := widget.NewLabelWithStyle(
		"Loading pastes...",
		fyne.TextAlignCenter,
		fyne.TextStyle{},
	)
	
	// Add the status label to the current container temporarily
	overlay := container.NewCenter(statusLabel)
	g.mainWindow.SetContent(container.NewStack(g.currentContainer, overlay))

	go func() {
		pastes, err := g.pasteApp.GetUserPastes()
		
		// Update UI in a goroutine-safe way
		g.mainWindow.Canvas().Refresh(g.currentContainer)
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to load pastes: %v", err), g.mainWindow)
			// Restore the original content
			g.mainWindow.SetContent(g.currentContainer)
			return
		}

		// Store pastes for access in the list
		pastesData := pastes

		// Update list data
		list.Length = func() int {
			return len(pastesData)
		}

		list.UpdateItem = func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < 0 || id >= len(pastesData) {
				return
			}

			paste := pastesData[id]
			border := obj.(*fyne.Container)
			
			// The title is in the content part of the border layout
			titleLabel := border.Objects[0].(*widget.Label)
			titleLabel.SetText(paste.Title)
			
			// The date is in the right part of the border layout
			dateLabel := border.Objects[1].(*widget.Label)
			dateLabel.SetText(paste.CreatedAt.Format(time.DateOnly))
		}

		// Set up tap handler
		list.OnSelected = func(id widget.ListItemID) {
			if id >= 0 && id < len(pastesData) {
				g.showPasteDetails(pastesData[id])
			}
			list.UnselectAll()
		}

		// Refresh the list
		list.Refresh()
		
		// Restore the original content
		g.mainWindow.SetContent(g.currentContainer)
	}()
}

// showPasteDetails shows the details of a selected paste
func (g *GUI) showPasteDetails(paste *models.Paste) {
	progress := dialog.NewProgress("Loading", "Decrypting paste...", g.mainWindow)
	progress.Show()

	go func() {
		title, content, err := g.pasteApp.GetPaste(paste.ID)
		
		// Update UI in a goroutine-safe way
		g.mainWindow.Canvas().Refresh(g.currentContainer)
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to decrypt paste: %v", err), g.mainWindow)
			progress.Hide()
			return
		}

		progress.Hide()

		contentView := container.NewVBox(
			widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewLabel(fmt.Sprintf("Created: %s", paste.CreatedAt.Format(time.RFC822))),
			widget.NewSeparator(),
			container.NewScroll(widget.NewLabel(content)),
		)

		dialog.ShowCustom("Paste Details", "Close", contentView, g.mainWindow)
	}()
}

// createNewPasteTab creates the tab for creating a new paste
func (g *GUI) createNewPasteTab() fyne.CanvasObject {
	// Create a styled heading
	heading := widget.NewLabelWithStyle(
		"Create a New Paste",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	// Create professional form fields with better styling
	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Enter paste title here...")

	contentEntry := widget.NewMultiLineEntry()
	contentEntry.SetPlaceHolder("Enter paste content here...")
	// Set a minimum size for better UX
	contentScroll := container.NewScroll(contentEntry)
	contentScroll.SetMinSize(fyne.NewSize(400, 300))

	// Add a styled checkbox for public visibility
	isPublicCheck := widget.NewCheck("Make paste public", nil)
	isPublicCheck.SetChecked(false)

	// Status label for showing creation progress
	statusLabel := widget.NewLabelWithStyle(
		"",
		fyne.TextAlignCenter,
		fyne.TextStyle{},
	)
	statusLabel.Hide()

	// Create buttons with better styling
	var createButton *widget.Button
	createButton = widget.NewButton("Create Paste", func() {
		title := strings.TrimSpace(titleEntry.Text)
		content := strings.TrimSpace(contentEntry.Text)

		if title == "" || content == "" {
			dialog.ShowError(fmt.Errorf("title and content are required"), g.mainWindow)
			return
		}

		// Show status in the app instead of progress dialog
		statusLabel.SetText("Creating your paste...")
		statusLabel.Show()
		createButton.Disable()

		go func() {
			// Call the correct method with all required parameters
			paste, err := g.pasteApp.CreatePaste(title, content, isPublicCheck.Checked, 0, 0)
			
			// Update UI in a goroutine-safe way
			g.mainWindow.Canvas().Refresh(g.currentContainer)
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("Failed to create paste: %v", err))
				createButton.Enable()
				return
			}

			statusLabel.Hide()
			createButton.Enable()
			dialog.ShowInformation("Success", fmt.Sprintf("Your paste has been created with ID: %s", paste.ID), g.mainWindow)
			
			// Clear the form
			titleEntry.SetText("")
			contentEntry.SetText("")
			isPublicCheck.SetChecked(false)
		}()
	})
	createButton.Importance = widget.HighImportance

	clearButton := widget.NewButton("Clear", func() {
		titleEntry.SetText("")
		contentEntry.SetText("")
		isPublicCheck.SetChecked(false)
	})

	// Create a more professional form layout with proper spacing
	form := container.NewVBox(
		heading,
		widget.NewSeparator(),
		container.NewPadded(
			container.NewVBox(
				widget.NewLabelWithStyle("Title", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				titleEntry,
				widget.NewLabelWithStyle("Content", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				contentScroll,
				isPublicCheck,
				statusLabel,
				container.NewHBox(
					layout.NewSpacer(),
					clearButton,
					createButton,
				),
			),
		),
	)

	return form
}

// createAccountTab creates the account settings tab
func (g *GUI) createAccountTab() fyne.CanvasObject {
	// Create a styled heading
	heading := widget.NewLabelWithStyle(
		"Account Information",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	// Create professional info labels with better styling
	emailLabel := widget.NewLabelWithStyle(
		"Email: Loading...",
		fyne.TextAlignLeading,
		fyne.TextStyle{},
	)

	createdAtLabel := widget.NewLabelWithStyle(
		"Account Created: Loading...",
		fyne.TextAlignLeading,
		fyne.TextStyle{},
	)

	// Create a styled security section
	securityHeading := widget.NewLabelWithStyle(
		"Security",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	securityInfo1 := widget.NewLabelWithStyle(
		"All your data is encrypted with zero-knowledge security.",
		fyne.TextAlignLeading,
		fyne.TextStyle{},
	)

	securityInfo2 := widget.NewLabelWithStyle(
		"Your password is never stored and is used to derive encryption keys.",
		fyne.TextAlignLeading,
		fyne.TextStyle{},
	)

	// Load user info asynchronously
	go func() {
		// Get user info from the current user
		if g.pasteApp.CurrentUser != nil {
			emailLabel.SetText(fmt.Sprintf("Email: %s", g.pasteApp.CurrentUser.Email))
			
			// Display the account creation date if available
			if !g.pasteApp.CurrentUser.CreatedAt.IsZero() {
				createdAtLabel.SetText(fmt.Sprintf("Account Created: %s", 
					g.pasteApp.CurrentUser.CreatedAt.Format(time.DateTime)))
			} else {
				createdAtLabel.SetText("Account Created: Information not available")
			}
		} else {
			emailLabel.SetText("Email: Not available")
			createdAtLabel.SetText("Account Created: Not available")
		}
		
		// Update UI in a goroutine-safe way
		g.mainWindow.Canvas().Refresh(g.currentContainer)
	}()

	// Create a card-like container for account info
	accountInfoCard := container.NewVBox(
		heading,
		widget.NewSeparator(),
		container.NewPadded(emailLabel),
		container.NewPadded(createdAtLabel),
	)

	// Create a card-like container for security info
	securityInfoCard := container.NewVBox(
		securityHeading,
		widget.NewSeparator(),
		container.NewPadded(securityInfo1),
		container.NewPadded(securityInfo2),
	)

	// Combine the cards with proper spacing
	form := container.NewVBox(
		accountInfoCard,
		widget.NewSeparator(),
		securityInfoCard,
	)

	return container.NewPadded(form)
}
