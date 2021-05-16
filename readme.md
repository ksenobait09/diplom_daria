# Как запускать
### bash
```bash
docker build -t daria . && docker run -p 1323:1323 -v $PWD:/daria
```

### cmd
```bash
docker build -t daria . && docker run -p 1323:1323 -v %cd%:/daria
```
