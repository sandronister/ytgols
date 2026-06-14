# ytgols

CLI em Go para baixar fluxos de vídeo ou áudio do YouTube.

> Use apenas em conteúdos que você tem autorização para baixar e respeite os
> termos do YouTube e a legislação aplicável.

## Execução

Requisitos:

- Go 1.26 ou superior;
- FFmpeg disponível no `PATH` para conversão de áudio.

```bash
go run ./cmd
```

O programa solicitará:

- URL do vídeo;
- tipo de mídia;
- qualidade;
- diretório de destino;
- nome do arquivo;
- `itag` opcional;
- metadados ID3v2.4 opcionais para áudio: título, artista, álbum, ano,
  gênero e número da faixa.

Se o diretório relativo já existir no local de execução, ele será usado. Caso
não exista e o local tenha um diretório pai, o destino será criado um nível
acima. Na raiz do sistema, o destino será criado no próprio local.

O modo de áudio converte o fluxo baixado para MP3. Para isso, o executável
`ffmpeg` precisa estar instalado e disponível no `PATH`. Pressione Enter para
aceitar os valores exibidos entre colchetes ou deixar metadados opcionais em
branco.
