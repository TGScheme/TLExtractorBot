import jadx.api.CommentsLevel;
import jadx.api.JadxArgs;
import jadx.api.JadxDecompiler;
import jadx.api.JavaClass;

import java.io.File;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;
import java.util.concurrent.atomic.AtomicInteger;

public final class TLExtract {
    public static void main(String[] argv) throws Exception {
        if (argv.length < 4) {
            System.err.println("usage: TLExtract <apk> <sources-dir> <package-prefix> <threads>");
            System.exit(2);
        }
        File apk = new File(argv[0]);
        Path sources = Path.of(argv[1]);
        String prefix = argv[2];
        int threads = Math.max(1, Integer.parseInt(argv[3]));

        JadxArgs args = new JadxArgs();
        args.setInputFiles(Collections.singletonList(apk));
        args.setOutDir(sources.toFile());
        args.setOutDirSrc(sources.toFile());
        args.setSkipResources(true);
        args.setReplaceConsts(false);
        args.setInlineAnonymousClasses(false);
        args.setCommentsLevel(CommentsLevel.NONE);
        args.setThreadsCount(threads);

        try (JadxDecompiler jadx = new JadxDecompiler(args)) {
            jadx.load();

            List<JavaClass> selected = new ArrayList<>();
            for (JavaClass cls : jadx.getClasses()) {
                if (cls.getFullName().startsWith(prefix)) {
                    selected.add(cls);
                }
            }
            if (selected.isEmpty()) {
                System.err.println("no class matched " + prefix);
                System.exit(1);
            }

            int total = selected.size();
            AtomicInteger done = new AtomicInteger();
            ExecutorService pool = Executors.newFixedThreadPool(threads);
            List<Future<?>> tasks = new ArrayList<>(total);
            for (JavaClass cls : selected) {
                tasks.add(pool.submit(() -> {
                    Path dir = sources.resolve(cls.getPackage().replace('.', File.separatorChar));
                    Files.createDirectories(dir);
                    String code = cls.getCode();
                    if (!code.endsWith("\n")) {
                        code = code + "\n";
                    }
                    Files.writeString(dir.resolve(cls.getName() + ".java"), code, StandardCharsets.UTF_8);
                    int at = done.incrementAndGet();
                    System.out.printf("INFO  - progress: %d of %d (%d%%)%n", at, total, at * 100 / total);
                    return null;
                }));
            }
            pool.shutdown();
            for (Future<?> task : tasks) {
                task.get();
            }
        }
    }
}
