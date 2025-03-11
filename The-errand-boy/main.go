package main

import (
	"fmt"
	"image"
	_ "image/png"
	"log"
	"math/rand"
	"time"

	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth     = 640
	screenHeight    = 480
	gridSize        = 32
	gridWidth       = screenWidth / gridSize
	gridHeight      = screenHeight / gridSize
	playerSpeed     = 5
	numLanes        = 8 // Количество полос
	numCarsPerLane  = 5 // Количество машин на полосу
	carSpeedMin     = 1.5
	carSpeedMax     = 3.0
	minCarGap       = 4              // Минимальный зазор между машинами
	maxCarGap       = 8              // Максимальный зазор между машинами
	initialGameTime = 60             // Начальное время игры в секундах
	laneSpacing     = gridSize * 1.5 // Расстояние между полосами
	textAreaHeight  = 50             // Высота области для текста
)

type GameObject struct {
	x, y    float64
	speed   float64
	image   *ebiten.Image
	width   int
	height  int
	isRight bool
}

type Game struct {
	player         *GameObject
	background     *ebiten.Image
	objects        map[string]*ebiten.Image
	cars           []*GameObject
	currentTime    int
	lastUpdateTime time.Time
	gameState      string
	elapsedTime    float64 // Накопленное время для отсчета секунд
}

func NewGame() *Game {
	g := &Game{
		currentTime:    initialGameTime,
		lastUpdateTime: time.Now(),
		objects:        make(map[string]*ebiten.Image),
		gameState:      "playing",
		player:         &GameObject{},
		elapsedTime:    0, // Инициализация накопленного времени
	}
	g.LoadImages()
	g.initializeGame()
	return g
}

func (g *Game) LoadImages() {
	// Загрузка изображений
	g.objects["car"] = LoadImage("car.png", 64, 32)                                  // Автомобиль 64x32
	g.objects["player"] = LoadImage("player.png", 32, 32)                            // Игрок 32x32
	g.objects["background"] = LoadImage("background.png", screenWidth, screenHeight) // Фон 640x480

	// Установка фона и изображения игрока
	g.background = g.objects["background"]
	g.player.image = g.objects["player"]
}

func LoadImage(path string, width, height int) *ebiten.Image {
	img, _, err := ebitenutil.NewImageFromFile(path)
	if err != nil {
		log.Printf("Failed to load image: %v, using placeholder", err)
		return ebiten.NewImage(width, height) // Заглушка
	}

	// Масштабирование изображения до нужных размеров
	scaledImg := ebiten.NewImage(width, height)
	op := &ebiten.DrawImageOptions{}

	// Рассчитываем масштаб
	origWidth, origHeight := img.Size()
	scaleX := float64(width) / float64(origWidth)
	scaleY := float64(height) / float64(origHeight)
	op.GeoM.Scale(scaleX, scaleY)

	scaledImg.DrawImage(img, op)
	return scaledImg
}

func (g *Game) initializeGame() {
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
		lastCarX := -float64(gridSize) // Стартовая позиция перед экраном
		for i := 0; i < numCarsPerLane; i++ {
			// Расчет зазора между автомобилями
			minGap := lastCarX + float64(minCarGap*gridSize)
			maxGap := lastCarX + float64(maxCarGap*gridSize)
			carX := minGap + rand.Float64()*(maxGap-minGap)

			g.cars = append(g.cars, &GameObject{
				x:       carX,
				y:       float64(lane)*laneSpacing + textAreaHeight, // Разные полосы для автомобилей
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
	if g.gameState != "playing" {
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			g.initializeGame()
			g.currentTime = initialGameTime
			g.gameState = "playing"
			g.elapsedTime = 0 // Сброс накопленного времени
		}
		return nil
	}

	now := time.Now()
	elapsed := now.Sub(g.lastUpdateTime)
	g.lastUpdateTime = now

	// Управление игроком
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		g.player.x -= gridSize * elapsed.Seconds() * playerSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		g.player.x += gridSize * elapsed.Seconds() * playerSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		g.player.y -= gridSize * elapsed.Seconds() * playerSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		g.player.y += gridSize * elapsed.Seconds() * playerSpeed
	}

	// Ограничение движения игрока в пределах экрана
	g.player.x = clamp(g.player.x, 0, screenWidth-float64(gridSize))
	g.player.y = clamp(g.player.y, textAreaHeight, screenHeight-float64(gridSize))

	// Проверка, достиг ли игрок верхней границы
	if g.player.y <= textAreaHeight {
		g.gameState = "win"
		return nil
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
	if g.elapsedTime >= 1.0 { // Если прошла целая секунда
		g.currentTime -= 1
		g.elapsedTime = 0 // Сброс накопленного времени
	}

	if g.currentTime <= 0 {
		g.gameState = "win"
	}

	return nil
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
		// Отрисовка плейсхолдера прямоугольной формы для игрока
		vector.DrawFilledRect(screen,
			float32(g.player.x),
			float32(g.player.y),
			float32(gridSize),
			float32(gridSize),
			color.RGBA{255, 0, 0, 255},
			false)
	}

	// Отрисовка времени
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Time: %d", g.currentTime), 10, 10)

	// Отрисовка состояния игры
	if g.gameState == "win" {
		ebitenutil.DebugPrintAt(screen, "You won! Press SPACE to restart.", 10, 30)
	} else if g.gameState == "lose" {
		ebitenutil.DebugPrintAt(screen, "Game over! Press SPACE to restart.", 10, 30)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("2D Game")

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
