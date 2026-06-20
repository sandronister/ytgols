package form

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Form builds and reads the interactive download form.
type Form struct {
	reader *bufio.Reader
	output io.Writer
}

// New creates an interactive form using the provided input and output.
func New(input io.Reader, output io.Writer) *Form {
	return &Form{
		reader: bufio.NewReader(input),
		output: output,
	}
}

// Ask displays the questions and returns the collected answers.
func (f *Form) Ask() (Answers, error) {
	fmt.Fprintln(f.output, "Download de vídeo ou áudio do YouTube")
	fmt.Fprintln(f.output)

	url, err := f.required("URL do vídeo")
	if err != nil {
		return Answers{}, err
	}

	mediaType, err := f.choice("Tipo", "video", "video", "audio")
	if err != nil {
		return Answers{}, err
	}

	quality, err := f.choice("Qualidade", "best", "best", "worst")
	if err != nil {
		return Answers{}, err
	}

	outputDir, err := f.text("Diretório de destino", "downloads")
	if err != nil {
		return Answers{}, err
	}

	filename, err := f.text("Nome do arquivo (opcional)", "")
	if err != nil {
		return Answers{}, err
	}

	itag, err := f.integer("Itag específico (opcional)", 0)
	if err != nil {
		return Answers{}, err
	}

	var metadata ID3Metadata
	if mediaType == "audio" {
		metadata, err = f.id3Metadata()
		if err != nil {
			return Answers{}, err
		}
	}

	return Answers{
		URL:       url,
		MediaType: mediaType,
		Quality:   quality,
		OutputDir: outputDir,
		Filename:  filename,
		Itag:      itag,
		Metadata:  metadata,
	}, nil
}

func (f *Form) id3Metadata() (ID3Metadata, error) {
	fmt.Fprintln(f.output)
	fmt.Fprintln(f.output, "Metadados ID3v2.4 (opcionais)")

	labels := []string{"Título", "Artista", "Álbum", "Ano", "Gênero", "Número da faixa"}
	values := make([]string, len(labels))
	for i, label := range labels {
		value, err := f.text(label, "")
		if err != nil {
			return ID3Metadata{}, err
		}
		values[i] = value
	}

	return ID3Metadata{
		Title:  values[0],
		Artist: values[1],
		Album:  values[2],
		Year:   values[3],
		Genre:  values[4],
		Track:  values[5],
	}, nil
}

func (f *Form) AskAgain() (bool, error) {
	answer, err := f.choice("Deseja fazer outro download", "n", "s", "n")
	if err != nil {
		return false, err
	}
	return answer == "s", nil
}

func (f *Form) text(label, defaultValue string) (string, error) {
	if defaultValue == "" {
		fmt.Fprintf(f.output, "%s: ", label)
	} else {
		fmt.Fprintf(f.output, "%s [%s]: ", label, defaultValue)
	}

	value, err := f.reader.ReadString('\n')
	if err != nil && len(value) == 0 {
		return "", err
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return defaultValue, nil
	}
	return value, nil
}

func (f *Form) required(label string) (string, error) {
	for {
		value, err := f.text(label, "")
		if err != nil {
			return "", err
		}
		if value != "" {
			return value, nil
		}
		fmt.Fprintln(f.output, "Este campo é obrigatório.")
	}
}

func (f *Form) choice(label, defaultValue string, choices ...string) (string, error) {
	for {
		value, err := f.text(label+" ("+strings.Join(choices, "/")+")", defaultValue)
		if err != nil {
			return "", err
		}
		value = strings.ToLower(value)
		for _, choice := range choices {
			if value == choice {
				return value, nil
			}
		}
		fmt.Fprintf(f.output, "Valor inválido. Escolha: %s.\n", strings.Join(choices, " ou "))
	}
}

func (f *Form) integer(label string, defaultValue int) (int, error) {
	for {
		value, err := f.text(label, "")
		if err != nil {
			return 0, err
		}
		if value == "" {
			return defaultValue, nil
		}

		number, conversionErr := strconv.Atoi(value)
		if conversionErr == nil && number >= 0 {
			return number, nil
		}
		fmt.Fprintln(f.output, "Informe um número inteiro maior ou igual a zero.")
	}
}
