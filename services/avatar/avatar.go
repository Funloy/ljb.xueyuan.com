// @APIVersion 1.0.0
// @Title 用户头像生成服务
// @Description 本服务用户随机生成用户头像
// @Contact xuchuangxin@icanmake.cn
// @TermsOfServiceUrl https://maiyajia.com/
// @License
// @LicenseUrl

package avatar

import (
	"bytes"
	"encoding/base64"
	"errors"
	"hash/fnv"
	"os"

	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math/rand"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/deiwin/picasso"
	"maiyajia.com/services/avatar/bindata"
)

var errUnknownGender = errors.New("Unknown gender")

const (
	avatarWidth  int = 200
	avatarHeight int = 200
)

type person struct {
	Clothes []string
	Eye     []string
	Face    []string
	Hair    []string
	Mouth   []string
}

type store struct {
	Background []string
	Male       person
	Female     person
}

var assetsStore *store

// Gender 性别
type Gender int

// 男女性别常量
const (
	MALE Gender = iota
	FEMALE
)

func init() {
	male := getPerson(MALE)
	female := getPerson(FEMALE)
	assetsStore = &store{Background: readAssetsFrom("data/background"), Male: male, Female: female}
	rand.Seed(time.Now().UTC().UnixNano())
}

// Builder 根据性别随机产生头像Image编码
func Builder(gender Gender) (image.Image, error) {
	switch gender {
	case MALE:
		return randomAvatar(assetsStore.Male, time.Now().UnixNano())
	case FEMALE:
		return randomAvatar(assetsStore.Female, time.Now().UnixNano())
	default:
		return nil, errUnknownGender
	}
}

// BuilderWithSalt 根据性别和种子字符串(可选)随机产生头像Image编码
func BuilderWithSalt(gender Gender, salt string) (image.Image, error) {
	h := fnv.New32a()
	_, err := h.Write([]byte(salt))
	if err != nil {
		return nil, err
	}
	switch gender {
	case MALE:
		return randomAvatar(assetsStore.Male, int64(h.Sum32()))
	case FEMALE:
		return randomAvatar(assetsStore.Female, int64(h.Sum32()))
	default:
		return nil, errUnknownGender
	}
}

// ImageToBase64 把png图片转换成base64字符串格式
func ImageToBase64(img image.Image) string {

	buf := new(bytes.Buffer)
	png.Encode(buf, img)
	imgBytes := buf.Bytes()

	//base64压缩，编码成字符串
	sourcestring := base64.StdEncoding.EncodeToString(imgBytes)

	return "data:image/png;base64," + sourcestring
}

// Base64Buidler 根据性别和种子字符串(可选)随机产生base64编码的png格式的头像
func Base64Buidler(gender Gender) (string, error) {

	img, err := Builder(gender)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	png.Encode(buf, img)
	imgBytes := buf.Bytes()

	//base64压缩，编码成字符串
	sourcestring := base64.StdEncoding.EncodeToString(imgBytes)

	return "data:image/png;base64," + sourcestring, err
}

// CompositeBuilder 根据传入的图片数组，把多张图片组合成一张头像
func CompositeBuilder(images []image.Image) image.Image {

	size := len(images)
	gray := color.RGBA{0xea, 0xea, 0xea, 0xff}
	var image image.Image

	switch {
	case size == 3:
		image = picasso.VerticalSplit{
			Ratio: 0.5,
			Right: picasso.Picture{Picture: images[0]},
			Left: picasso.HorizontalSplit{
				Ratio:  1,
				Top:    picasso.Picture{Picture: images[1]},
				Bottom: picasso.Picture{Picture: images[2]},
			},
		}.DrawWithBorder(400, 400, gray, 15)
		break
	case size == 4:
		image = picasso.HorizontalSplit{
			Ratio: 2,
			Top:   picasso.Picture{Picture: images[0]},
			Bottom: picasso.VerticalSplit{
				Ratio: 0.5,
				Left:  picasso.Picture{Picture: images[1]},
				Right: picasso.VerticalSplit{
					Ratio: 1,
					Left:  picasso.Picture{Picture: images[2]},
					Right: picasso.Picture{Picture: images[3]},
				},
			},
		}.DrawWithBorder(400, 400, gray, 15)
		break
	default:
		layout := picasso.GoldenSpiralLayout()
		image = layout.Compose(images).DrawWithBorder(400, 400, gray, 5)
	}

	return image

}

// SaveAvatarToFile 把头像保存到文件系统
func SaveAvatarToFile(img image.Image, filePath string) error {
	outFile, err := os.Create(filePath)
	defer outFile.Close()
	if err != nil {
		return err
	}
	switch strings.ToLower(filepath.Ext(filePath)) {
	case ".jpeg", ".jpg":
		err = jpeg.Encode(outFile, img, &jpeg.Options{Quality: 80})
	case ".gif":
		err = gif.Encode(outFile, img, nil)
	default:
		err = png.Encode(outFile, img)
	}
	return err
}

// ReadAvatarFromFile 把头像保存到文件系统
func ReadAvatarFromFile(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	image, _, err := image.Decode(file)
	return image, err

}

func randomAvatar(p person, seed int64) (image.Image, error) {
	rnd := rand.New(rand.NewSource(seed))
	avatar := image.NewRGBA(image.Rect(0, 0, avatarWidth, avatarHeight))
	var err error
	err = drawImg(avatar, randSliceString(rnd, assetsStore.Background), err)
	err = drawImg(avatar, randSliceString(rnd, p.Face), err)
	err = drawImg(avatar, randSliceString(rnd, p.Clothes), err)
	err = drawImg(avatar, randSliceString(rnd, p.Mouth), err)
	err = drawImg(avatar, randSliceString(rnd, p.Hair), err)
	err = drawImg(avatar, randSliceString(rnd, p.Eye), err)
	return avatar, err
}

func drawImg(dst draw.Image, asset string, err error) error {
	if err != nil {
		return err
	}
	src, _, err := image.Decode(bytes.NewReader(bindata.MustAsset(asset)))
	if err != nil {
		return err
	}
	draw.Draw(dst, dst.Bounds(), src, image.Point{0, 0}, draw.Over)

	return nil
}

func getPerson(gender Gender) person {
	var genderPath string

	switch gender {
	case FEMALE:
		genderPath = "female"
	case MALE:
		genderPath = "male"
	}

	return person{
		Clothes: readAssetsFrom("data/" + genderPath + "/clothes"),
		Eye:     readAssetsFrom("data/" + genderPath + "/eye"),
		Face:    readAssetsFrom("data/" + genderPath + "/face"),
		Hair:    readAssetsFrom("data/" + genderPath + "/hair"),
		Mouth:   readAssetsFrom("data/" + genderPath + "/mouth"),
	}
}

func readAssetsFrom(dir string) []string {
	assets, _ := bindata.AssetDir(dir)
	for i, asset := range assets {
		assets[i] = filepath.Join(dir, asset)
	}
	sort.Sort(naturalSort(assets))
	return assets
}
