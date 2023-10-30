from typing import List
from fastapi import FastAPI
from pydantic import BaseModel
from analyzer import Analyzer
import re

import gensim.downloader
from nltk import pos_tag
from pymorphy2 import MorphAnalyzer

SIMILARITY_THRESHOLD = 0.7

morph = MorphAnalyzer()
model_ru = gensim.downloader.load("word2vec-ruscorpora-300")

bad_letters = "[0-9!#$%&'()*+,./:;<=>?@[\]^_`{|}~—\"\-]+"


def make_analyzer() -> Analyzer:
    def split_func(text):
        text = re.sub(bad_letters, "", text)
        return text.split()

    def normalize(a, lang):
        # Get normal form.
        a = morph.normal_forms(a.strip())[0]
        # Get tag.
        a, tag = pos_tag([a], tagset="universal", lang=lang)[0]
        return f"{a}_{tag}"

    def recognize_func(a, b):
        arus = normalize(a, lang="rus")
        brus = normalize(b, lang="rus")

        try:
            similarity = model_ru.similarity(arus, brus)
            print(similarity)
            return similarity
        except Exception as e:
            print(e)

        if a.lower() == b.lower():
            return 1
        return 0

    return Analyzer(split_func, recognize_func, SIMILARITY_THRESHOLD)


class AnalyzeQuery(BaseModel):
    topics: List[str]
    text: str


class AnalyzeResponse(BaseModel):
    topics: List[str]


app = FastAPI()
analyzer = make_analyzer()


@app.post("/analyze")
def analyze(query: AnalyzeQuery) -> AnalyzeResponse:
    return AnalyzeResponse(topics=list(
        filter(lambda t: analyzer.contain_topic(query.text, t), query.topics)))
