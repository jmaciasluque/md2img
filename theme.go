package md2img

import "fmt"

// ApplyTheme applies a named visual preset to cfg.
func ApplyTheme(cfg *Config, name string) error {
	switch name {
	case "", "light":
		return nil
	case "dark":
		cfg.TextColor = Color{226, 232, 240}
		cfg.LinkColor = Color{147, 197, 253}
		cfg.HeadingColor = Color{248, 250, 252}
		cfg.TableHeaderBg = Color{30, 41, 59}
		cfg.TableHeaderFg = Color{248, 250, 252}
		cfg.TableRowEven = Color{15, 23, 42}
		cfg.TableRowOdd = Color{30, 41, 59}
		cfg.CodeBg = Color{15, 23, 42}
		cfg.BlockquoteLineColor = Color{96, 165, 250}
		cfg.BlockquoteTextColor = Color{203, 213, 225}
		cfg.HRColor = Color{71, 85, 105}
	case "github":
		cfg.TextColor = Color{31, 35, 40}
		cfg.LinkColor = Color{9, 105, 218}
		cfg.HeadingColor = Color{31, 35, 40}
		cfg.TableHeaderBg = Color{246, 248, 250}
		cfg.TableHeaderFg = Color{31, 35, 40}
		cfg.TableRowEven = Color{246, 248, 250}
		cfg.TableRowOdd = Color{255, 255, 255}
		cfg.CodeBg = Color{246, 248, 250}
		cfg.BlockquoteLineColor = Color{208, 215, 222}
		cfg.BlockquoteTextColor = Color{87, 96, 106}
		cfg.HRColor = Color{208, 215, 222}
	case "slack":
		cfg.TextColor = Color{29, 28, 29}
		cfg.LinkColor = Color{18, 100, 163}
		cfg.HeadingColor = Color{29, 28, 29}
		cfg.TableHeaderBg = Color{74, 21, 75}
		cfg.TableHeaderFg = Color{255, 255, 255}
		cfg.TableRowEven = Color{248, 248, 248}
		cfg.TableRowOdd = Color{255, 255, 255}
		cfg.CodeBg = Color{244, 244, 244}
		cfg.BlockquoteLineColor = Color{221, 221, 221}
		cfg.BlockquoteTextColor = Color{97, 96, 97}
		cfg.HRColor = Color{221, 221, 221}
	default:
		return fmt.Errorf("unknown theme %q", name)
	}
	return nil
}
