package main

import (
	"fmt"
	"image"
	_ "image/png"
	"log"
	"math/rand"
	"os"
	"time"

	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

const (
	screenWidth    = 640
	screenHeight   = 480
	gridSize       = 32
	gridWidth      = screenWidth / gridSize
	gridHeight     = screenHeight / gridSize
	playerSpeed    = 5
	laneSpacing    = gridSize * 1.5 // Расстояние между полосами
	textAreaHeight = 5             // Высота области для текста
)

// Константы для уровней сложности
const (
	Easy = iota
	Medium
	Hard
)

type GameObject struct {
	x, y    float64
	speed   float64
	image   *ebiten.Image
	width   int
	height  int
	isRight bool
}

type Button struct {
	x, y, width, height float64
	text                string
	hovered             bool
	action              func()
}

type Game struct {
	player         *GameObject
	background     *ebiten.Image
	objects        map[string]*ebiten.Image
	cars           []*GameObject
	currentTime    int
	lastUpdateTime time.Time
	gameState      string // "menu", "playing", "paused", "win", "lose"
	elapsedTime    float64
	font           font.Face
	buttons        map[string]*Button
	difficulty     int
	levelTime      int // Время для текущего уровня
}

func NewGame() *Game {
	g := &Game{
		objects:     make(map[string]*ebiten.Image),
		gameState:   "menu",
		player:      &GameObject{},
		font:        basicfont.Face7x13,
		buttons:     make(map[string]*Button),
		difficulty:  Easy, // Начинаем с легкого уровня
	}
	g.LoadImages()
	g.createButtons()
	g.setDifficulty(Easy) // Устанавливаем начальную сложность
	return g
}

// Установка параметров сложности
func (g *Game) setDifficulty(level int) {
	g.difficulty = level
	
	switch level {
	case Easy:
		g.levelTime = 30 
	case Medium:
		g.levelTime = 25 
	case Hard:
		g.levelTime = 10 
	}
	
	g.currentTime = g.levelTime
	g.initializeGame()
}

// Получение параметров для текущего уровня сложности
func (g *Game) getLevelParams() (int, int, float64, float64, int, int) {
	switch g.difficulty {
	case Easy:
		return 7, 2, 2, 2.5, 7, 11 
	case Medium:
		return 8, 3, 2.5, 3.5, 6, 9  
	case Hard:
		return 9, 4, 3, 4.5, 5, 7  
	default:
		return 7, 2, 2, 2.5, 7, 11
	}
}

func (g *Game) createButtons() {
	// Кнопки выбора уровня сложности
	g.buttons["easy"] = &Button{
		x:      screenWidth/2 - 150,
		y:      screenHeight/2 - 20,
		width:  140,
		height: 40,
		text:   "Easy",
		action: func() {
			g.setDifficulty(Easy)
			g.gameState = "playing"
			g.lastUpdateTime = time.Now()
		},
	}

	g.buttons["medium"] = &Button{
		x:      screenWidth/2 - 70,
		y:      screenHeight/2 - 20,
		width:  140,
		height: 40,
		text:   "Medium",
		action: func() {
			g.setDifficulty(Medium)
			g.gameState = "playing"
			g.lastUpdateTime = time.Now()
		},
	}

	g.buttons["hard"] = &Button{
		x:      screenWidth/2 + 10,
		y:      screenHeight/2 - 20,
		width:  140,
		height: 40,
		text:   "Hard",
		action: func() {
			g.setDifficulty(Hard)
			g.gameState = "playing"
			g.lastUpdateTime = time.Now()
		},
	}

	// Кнопка "Выйти из игры" в главном меню
	g.buttons["exit_menu"] = &Button{
		x:      screenWidth/2 - 100,
		y:      screenHeight/2 + 40,
		width:  200,
		height: 40,
		text:   "Exit Game",
		action: func() {
			os.Exit(0)
		},
	}

	// Кнопка "Выйти в меню" в паузе
	g.buttons["exit_pause"] = &Button{
		x:      screenWidth/2 - 100,
		y:      screenHeight/2 + 50,
		width:  200,
		height: 40,
		text:   "Back to Menu",
		action: func() {
			g.gameState = "menu"
		},
	}

	// Кнопка "Рестарт" при проигрыше/выигрыше
	g.buttons["restart"] = &Button{
		x:      screenWidth/2 - 100,
		y:      screenHeight/2 + 50,
		width:  200,
		height: 40,
		text:   "Play Again",
		action: func() {
			g.setDifficulty(g.difficulty)
			g.gameState = "playing"
			g.lastUpdateTime = time.Now()
		},
	}

	// Кнопка "Меню" при проигрыше/выигрыше
	g.buttons["menu"] = &Button{
		x:      screenWidth/2 - 100,
		y:      screenHeight/2 + 100,
		width:  200,
		height: 40,
		text:   "Main Menu",
		action: func() {
			g.gameState = "menu"
		},
	}
}

func (g *Game) LoadImages() {
	// Загрузка реальных изображений вместо цветных placeholder'ов
	var err error
	
	// Загрузка изображения машины
	g.objects["car"], err = loadImageFromFile("bus.png", 64, 32)
	if err != nil {
		log.Printf("Failed to load car image: %v, using placeholder", err)
		g.objects["car"] = g.createPlaceholderImage(64, 32, color.RGBA{255, 0, 0, 255})
	}
	
	// Загрузка изображения игрока
	g.objects["player"], err = loadImageFromFile("player.png", 32, 32)
	if err != nil {
		log.Printf("Failed to load player image: %v, using placeholder", err)
		g.objects["player"] = g.createPlaceholderImage(32, 32, color.RGBA{0, 255, 0, 255})
	}
	
	// Загрузка фонового изображения
	g.objects["background"], err = loadImageFromFile("back.png", screenWidth, screenHeight)
	if err != nil {
		log.Printf("Failed to load background image: %v, using placeholder", err)
		g.objects["background"] = g.createPlaceholderImage(screenWidth, screenHeight, color.RGBA{200, 200, 200, 255})
	}

	g.background = g.objects["background"]
	g.player.image = g.objects["player"]
}

func loadImageFromFile(path string, targetWidth, targetHeight int) (*ebiten.Image, error) {
	img, _, err := ebitenutil.NewImageFromFile(path)
	if err != nil {
		return nil, err
	}
	
	origWidth, origHeight := img.Size()
	targetImg := ebiten.NewImage(targetWidth, targetHeight)
	op := &ebiten.DrawImageOptions{}
	
	scaleX := float64(targetWidth) / float64(origWidth)
	scaleY := float64(targetHeight) / float64(origHeight)
	op.GeoM.Scale(scaleX, scaleY)
	
	targetImg.DrawImage(img, op)
	return targetImg, nil
}

func (g *Game) createPlaceholderImage(width, height int, col color.RGBA) *ebiten.Image {
	img := ebiten.NewImage(width, height)
	img.Fill(col)
	return img
}

func (g *Game) initializeGame() {
	// Получаем параметры для текущего уровня сложности
	numLanes, numCarsPerLane, carSpeedMin, carSpeedMax, minCarGap, maxCarGap := g.getLevelParams()

	// Настройка начального положения игрока
	g.player = &GameObject{
		x:      float64(gridWidth/2) * gridSize,
		y:      float64((gridHeight - 1) * gridSize),
		speed:  playerSpeed,
		image:  g.objects["player"],
		width:  gridSize,
		height: gridSize,
	}

	// Очистка существующих автомобилей
	g.cars = []*GameObject{}

	// Инициализация автомобилей на каждой полосе
	for lane := 0; lane < numLanes; lane++ {
		lastCarX := -float64(gridSize)
		for i := 0; i < numCarsPerLane; i++ {
			minGap := lastCarX + float64(minCarGap*gridSize)
			maxGap := lastCarX + float64(maxCarGap*gridSize)
			carX := minGap + rand.Float64()*(maxGap-minGap)

			g.cars = append(g.cars, &GameObject{
				x:       carX,
				y:       float64(lane)*laneSpacing + textAreaHeight,
				speed:   carSpeedMin + rand.Float64()*(carSpeedMax-carSpeedMin),
				image:   g.objects["car"],
				width:   gridSize * 2,
				height:  gridSize,
				isRight: rand.Intn(2) == 0,
			})

			lastCarX = carX
		}
	}
}

func (g *Game) Update() error {
	// Обработка паузы
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) && (g.gameState == "playing" || g.gameState == "paused") {
		if g.gameState == "playing" {
			g.gameState = "paused"
		} else {
			g.gameState = "playing"
			g.lastUpdateTime = time.Now()
		}
	}

	switch g.gameState {
	case "menu":
		g.updateMenu()
	case "playing":
		g.updateGame()
	case "paused":
		g.updatePaused()
	case "win", "lose":
		g.updateGameOver()
	}

	return nil
}

func (g *Game) updateMenu() {
	// Обновление кнопок в меню
	mx, my := ebiten.CursorPosition()
	g.buttons["easy"].hovered = g.buttons["easy"].contains(float64(mx), float64(my))
	g.buttons["medium"].hovered = g.buttons["medium"].contains(float64(mx), float64(my))
	g.buttons["hard"].hovered = g.buttons["hard"].contains(float64(mx), float64(my))
	g.buttons["exit_menu"].hovered = g.buttons["exit_menu"].contains(float64(mx), float64(my))

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if g.buttons["easy"].hovered {
			g.buttons["easy"].action()
		}
		if g.buttons["medium"].hovered {
			g.buttons["medium"].action()
		}
		if g.buttons["hard"].hovered {
			g.buttons["hard"].action()
		}
		if g.buttons["exit_menu"].hovered {
			g.buttons["exit_menu"].action()
		}
	}
}

func (g *Game) updateGame() {
	now := time.Now()
	elapsed := now.Sub(g.lastUpdateTime)
	g.lastUpdateTime = now

	// Управление игроком
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.player.x -= gridSize * elapsed.Seconds() * playerSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.player.x += gridSize * elapsed.Seconds() * playerSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		g.player.y -= gridSize * elapsed.Seconds() * playerSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.player.y += gridSize * elapsed.Seconds() * playerSpeed
	}

	// Ограничение движения игрока
	g.player.x = clamp(g.player.x, 0, screenWidth-float64(gridSize))
	g.player.y = clamp(g.player.y, textAreaHeight, screenHeight-float64(gridSize))

	// Проверка победы - достиг верха экрана
	if g.player.y <= textAreaHeight {
		g.gameState = "win"
		return
	}

	// Обновление автомобилей
	for _, car := range g.cars {
		if car.isRight {
			car.x += car.speed * elapsed.Seconds() * gridSize
			if car.x > screenWidth {
				car.x = -float64(car.width)
			}
		} else {
			car.x -= car.speed * elapsed.Seconds() * gridSize
			if car.x < -float64(car.width) {
				car.x = screenWidth
			}
		}
	}

	// Проверка столкновений
	g.checkCollisions()

	// Обновление времени
	g.elapsedTime += elapsed.Seconds()
	if g.elapsedTime >= 1.0 {
		g.currentTime -= 1
		g.elapsedTime = 0
		
		// Проверка на проигрыш по времени
		if g.currentTime <= 0 {
			g.gameState = "lose" // Время вышло - проигрыш
			return
		}
	}
}

func (g *Game) updatePaused() {
	// Обновление кнопок в паузе
	mx, my := ebiten.CursorPosition()
	g.buttons["exit_pause"].hovered = g.buttons["exit_pause"].contains(float64(mx), float64(my))

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if g.buttons["exit_pause"].hovered {
			g.buttons["exit_pause"].action()
		}
	}
}

func (g *Game) updateGameOver() {
	// Обновление кнопок
	mx, my := ebiten.CursorPosition()
	g.buttons["restart"].hovered = g.buttons["restart"].contains(float64(mx), float64(my))
	g.buttons["menu"].hovered = g.buttons["menu"].contains(float64(mx), float64(my))

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if g.buttons["restart"].hovered {
			g.buttons["restart"].action()
		}
		if g.buttons["menu"].hovered {
			g.buttons["menu"].action()
		}
	}
}

func (g *Game) checkCollisions() {
	playerRect := image.Rect(int(g.player.x), int(g.player.y), int(g.player.x)+gridSize, int(g.player.y)+gridSize)

	for _, car := range g.cars {
		carRect := image.Rect(int(car.x), int(car.y), int(car.x)+car.width, int(car.y)+car.height)
		if playerRect.Overlaps(carRect) {
			g.gameState = "lose"
			return
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Отрисовка фона
	if g.background != nil {
		op := &ebiten.DrawImageOptions{}
		screen.DrawImage(g.background, op)
	}

	switch g.gameState {
	case "menu":
		g.drawMenu(screen)
	case "playing":
		g.drawGame(screen)
	case "paused":
		g.drawGame(screen)
		g.drawPauseMenu(screen)
	case "win", "lose":
		g.drawGame(screen)
		g.drawGameOver(screen)
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	// Фон меню с легкой прозрачностью
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{0, 0, 0, 180}, false)

	// Заголовок игры
	title := "ROAD ADVENTURE"
	titleBounds := text.BoundString(g.font, title)
	text.Draw(screen, title, g.font, screenWidth/2-titleBounds.Max.X/2, 80, color.RGBA{255, 215, 0, 255}) // Золотой цвет

	// Подзаголовок выбора сложности
	diffText := "SELECT DIFFICULTY"
	diffBounds := text.BoundString(g.font, diffText)
	text.Draw(screen, diffText, g.font, screenWidth/2-diffBounds.Max.X/2, 140, color.White)

	// Расположение кнопок сложности в одну линию с равными промежутками
	buttonY := 180
	g.buttons["easy"].x = screenWidth/2 - 220
	g.buttons["easy"].y = float64(buttonY)
	g.buttons["medium"].x = screenWidth/2 - 70
	g.buttons["medium"].y = float64(buttonY)
	g.buttons["hard"].x = screenWidth/2 + 80
	g.buttons["hard"].y = float64(buttonY)

	// Отрисовка кнопок сложности
	g.drawButton(screen, g.buttons["easy"])
	g.drawButton(screen, g.buttons["medium"])
	g.drawButton(screen, g.buttons["hard"])

	// Разделительная линия
	separatorY := buttonY + 70
	vector.StrokeLine(screen, screenWidth/4, float32(separatorY), screenWidth*3/4, float32(separatorY), 2, color.RGBA{100, 100, 100, 255}, false)

	// Управление
	controlsTitle := "CONTROLS"
	controlsTitleBounds := text.BoundString(g.font, controlsTitle)
	text.Draw(screen, controlsTitle, g.font, screenWidth/2-controlsTitleBounds.Max.X/2, separatorY+30, color.White)

	controls := []string{
		"W/A/S/D or Arrow Keys - Movement",
		"Space - Pause/Resume",
		"ESC - Back to Menu",
	}
	for i, line := range controls {
		lineBounds := text.BoundString(g.font, line)
		text.Draw(screen, line, g.font, screenWidth/2-lineBounds.Max.X/2, separatorY+60+i*20, color.RGBA{200, 200, 200, 255})
	}

	// Кнопка выхода - центрированная внизу
	g.buttons["exit_menu"].x = screenWidth/2 - 100
	g.buttons["exit_menu"].y = float64(screenHeight - 80)
	g.drawButton(screen, g.buttons["exit_menu"])

	// Версия игры или авторские права
	versionText := "v1.0 © 2024"
	versionBounds := text.BoundString(g.font, versionText)
	text.Draw(screen, versionText, g.font, screenWidth-versionBounds.Max.X-10, screenHeight-20, color.RGBA{150, 150, 150, 255})
}

func (g *Game) drawButton(screen *ebiten.Image, btn *Button) {
	// Цвет кнопки - градиент от синего к более светлому
	btnColor := color.RGBA{65, 105, 225, 255} // Royal Blue
	if btn.hovered {
		btnColor = color.RGBA{100, 149, 237, 255} // Cornflower Blue
	}

	// Основной прямоугольник кнопки
	vector.DrawFilledRect(screen, float32(btn.x), float32(btn.y), float32(btn.width), float32(btn.height), btnColor, false)
	
	// Тонкая рамка вокруг кнопки
	borderColor := color.RGBA{255, 255, 255, 100}
	vector.StrokeRect(screen, float32(btn.x), float32(btn.y), float32(btn.width), float32(btn.height), 1, borderColor, false)
	
	// Эффект тени при наведении
	if btn.hovered {
		shadowColor := color.RGBA{255, 255, 255, 50}
		vector.DrawFilledRect(screen, float32(btn.x)+2, float32(btn.y)+2, float32(btn.width), float32(btn.height), shadowColor, false)
	}

	// Текст кнопки
	textBounds := text.BoundString(g.font, btn.text)
	textHeight := g.font.Metrics().Height.Ceil()
	textColor := color.White
	
	// Полужирный эффект для текста при наведении
	if btn.hovered {
		textColor = color.Black 
	}
	
	text.Draw(screen, btn.text, g.font, 
		int(btn.x)+int(btn.width)/2-textBounds.Max.X/2, 
		int(btn.y)+int(btn.height)/2+textHeight/2-2, 
		textColor)
}

func (g *Game) drawGame(screen *ebiten.Image) {
	// Отрисовка автомобилей
	for _, car := range g.cars {
		if car.image != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(car.x, car.y)
			screen.DrawImage(car.image, op)
		}
	}

	// Отрисовка игрока
	if g.player.image != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(g.player.x, g.player.y)
		screen.DrawImage(g.player.image, op)
	} else {
		vector.DrawFilledRect(screen,
			float32(g.player.x),
			float32(g.player.y),
			float32(gridSize),
			float32(gridSize),
			color.RGBA{255, 0, 0, 255},
			false)
	}

	// Отрисовка времени и уровня сложности
	levelText := ""
	switch g.difficulty {
	case Easy:
		levelText = "Easy"
	case Medium:
		levelText = "Medium"
	case Hard:
		levelText = "Hard"
	}
	
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Time: %d", g.currentTime), 10, 10)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Level: %s", levelText), 10, 30)
}

func (g *Game) drawPauseMenu(screen *ebiten.Image) {
	// Полупрозрачный фон
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{0, 0, 0, 150}, false)

	// Текст паузы
	pauseText := "PAUSE"
	pauseBounds := text.BoundString(g.font, pauseText)
	text.Draw(screen, pauseText, g.font, screenWidth/2-pauseBounds.Max.X/2, screenHeight/2-50, color.White)

	// Кнопка "Выйти в меню"
	g.drawButton(screen, g.buttons["exit_pause"])
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	// Полупрозрачный фон
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{0, 0, 0, 150}, false)

	// Текст результата
	resultText := ""
	if g.gameState == "win" {
		resultText = "VICTORY!"
	} else {
		resultText = "GAME OVER!"
	}

	reasonText := ""
	if g.gameState == "lose" && g.currentTime <= 0 {
		reasonText = "Time's up!"
	} else if g.gameState == "lose" {
		reasonText = "You got hit!"
	} else {
		reasonText = "You made it!"
	}

	resultBounds := text.BoundString(g.font, resultText)
	text.Draw(screen, resultText, g.font, screenWidth/2-resultBounds.Max.X/2, screenHeight/2-80, color.White)
	
	reasonBounds := text.BoundString(g.font, reasonText)
	text.Draw(screen, reasonText, g.font, screenWidth/2-reasonBounds.Max.X/2, screenHeight/2-50, color.White)

	// Кнопки
	g.drawButton(screen, g.buttons["restart"])
	g.drawButton(screen, g.buttons["menu"])
}


func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("ROAD ADVENTURE")

	rand.Seed(time.Now().UnixNano())

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func (b *Button) contains(x, y float64) bool {
	return x >= b.x && x <= b.x+b.width && y >= b.y && y <= b.y+b.height
}