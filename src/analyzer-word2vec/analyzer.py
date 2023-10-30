
class Analyzer:
    # 0 <= threshold <= 1.
    def __init__(self, split_func, recognize_func, threshold) -> None:
        self.split_func = split_func
        self.recognize_func = recognize_func
        self.threshold = threshold
        pass

    def contain_topic(self, text, topic):
        print(text, topic)
        text_words = self.split_func(text)
        topic_words = self.split_func(topic)
        le = len(topic_words)
        for i in range(len(text_words) - le + 1):
            success = True
            for j in range(len(topic_words)):
                if (self.recognize_func(text_words[i+j], topic_words[j])
                        < self.threshold):
                    success = False
                    break
            if success:
                return True
        return False
