# calculadora http em Go
# estagio 1 (builder): imagem com o compilador Go (~250 MB) usada so para compilar
# estagio 2 (final)  : imagem "scratch" (vazia, 0 MB) contendo apenas o biaário
# resultado: imagem final com poucos MB sem shell e sem gerenciador de pacotes

# estagio1 - build
# versao fixada (nao "latest") => builds reproduziveis / variante alpine => imagem de build menor e download mais rapido
FROM golang:1.23-alpine AS builder

# diretorio de trabalho dentro do container de build
WORKDIR /src

# copia primeiro so go.mod e baixa dependencias / assim essa camada fica em cache e so eh refeita quando as dependencias mudam e nao a cada alteracao no codigo-fonte.
COPY go.mod ./
RUN go mod download

# agora copia o codigo (camada que muda com frequencia fica por ultimo)
COPY main.go ./

# compilaco
#  CGO_ENABLED=0 - binario estatico, sem depender da libc (obrigatorio p/ scratch) /  GOOS=linux   - sistema-alvo
#  -trimpath     - remove caminhos locais do binarios (reprodutibilidade/seguranca) /  -ldflags "-s -w" - remove tabela de simbulos e info de debug (binario ~30% menor)
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/calculadora .

# estagio2 - imagem final
# "scratch" eh uma imagem vazia: nao tem SO shell nem ferramentas / menor tamanho posivel e menor superficie de ataque
FROM scratch

# metadados da imagem (boas práticas OCI)
LABEL org.opencontainers.image.title="calculadora" \
      org.opencontainers.image.description="Calculadora HTTP em Go - multi-stage build"

# copia apenas o binario compilado do estagio anterior / compilador codigo-fonte e cache ficam para tras
COPY --from=builder /out/calculadora /calculadora

# executa como usuario sem privilegios (nobody) nunca como root
USER 65534:65534

# porta usada pela aplicacao (documentacao o mapeamento eh feito no docker run)
ENV PORT=8080
EXPOSE 8080

# healthcheck usando o proprio binario (scratch nao tem curl/wget).
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ["/calculadora", "-healthcheck"]

# forma "exec" (JSON): o processo vira PID 1 e recebe sinais (docker stop) corretamente
ENTRYPOINT ["/calculadora"]
