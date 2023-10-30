Установить библиотеки, указанные в `requirements.py`
```
pip install -r requirements.py
```

Дальше запустить приложение `make`.

При первом запуске выведенная ошибка подскажет, какие предтренированные модельки надо скачать (и как это сделать),
чтобы приложение заработало.

Также необходимо положить пустой файл `ru-rnc-new.map` в `$HOME/nltk_data/taggers/universal_tagset`
(https://github.com/nltk/nltk/issues/2434).