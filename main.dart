import 'package:flutter/material.dart';
import 'package:animated_text_kit/animated_text_kit.dart';

void main() => runApp(const RVNApp());

class RVNApp extends StatelessWidget {
  const RVNApp({super.key});
  @override Widget build(BuildContext context) => MaterialApp(
    debugShowCheckedModeBanner: false,
    theme: ThemeData(
      scaffoldBackgroundColor: Colors.black,
      fontFamily: 'Matrix',
      textTheme: const TextTheme(bodyMedium: TextStyle(color: Colors.lime)),
    ),
    home: const RVNScreen(),
  );
}

class RVNScreen extends StatefulWidget {
  const RVNScreen({super.key});
  @override State<RVNScreen> createState() => _RVNScreenState();
}

class _RVNScreenState extends State<RVNScreen> {
  bool connected = false;
  String currentCity = "Dallas";
  String currentIP = "10.66.66.10";
  String currentMAC = "02:42:AC:12:00:01";
  final List<String> threats = [];

  final cities = ["Dallas", "Houston", "New York", "Los Angeles", "London", "Tokyo", "Dubai", "Moscow", "São Paulo", "Lagos", "Sydney", "Berlin"];

  void connect() async {
    setState(() => connected = true);
    threats.add("${DateTime.now().toIso8601String().substring(11,19)} │ DPI probe blocked");
    threats.add("${DateTime.now().toIso8601String().substring(11,19)} │ Botnet C2 → 185.220.101.12 → dropped");
  }

  @override Widget build(BuildContext context) => Scaffold(
    body: Stack(
      children: [
        // Matrix rain background
        const MatrixRain(),
        SafeArea(
          child: Padding(
            padding: const EdgeInsets.all(20),
            child: Column(
              children: [
                AnimatedTextKit(
                  animatedTexts: [TyperAnimatedText('REALER VIRTUAL NETWORK', textStyle: const TextStyle(fontSize: 28, color: Colors.lime, fontWeight: FontWeight.bold))],
                  isRepeatingAnimation: false,
                ),
                const SizedBox(height: 30),
                // Connect button
                GestureDetector(
                  onTap: () => setState(() => connected = !connected),
                  child: Container(
                    width: double.infinity,
                    padding: const EdgeInsets.symmetric(vertical: 30),
                    decoration: BoxDecoration(color: connected ? Colors.lime : Colors.black, border: Border.all(color: Colors.lime, width: 3), borderRadius: BorderRadius.circular(15)),
                    child: Text(connected ? "● CONNECTED" : "▶ CONNECT", textAlign: TextAlign.center, style: TextStyle(fontSize: 36, color: connected ? Colors.black : Colors.lime, fontWeight: FontWeight.bold)),
                  ),
                ),
                const SizedBox(height: 30),
                // City picker
                DropdownButton<String>(
                  value: currentCity,
                  dropdownColor: Colors.black,
                  style: const TextStyle(color: Colors.lime, fontSize: 20),
                  items: cities.map((c) => DropdownMenuItem(value: c, child: Text("Exit → $c"))).toList(),
                  onChanged: (v) => setState(() => currentCity = v!),
                ),
                const SizedBox(height: 20),
                Text("IP → $currentIP", style: const TextStyle(fontSize: 20, color: Colors.lime)),
                Text("MAC → $currentMAC", style: const TextStyle(fontSize: 20, color: Colors.lime)),
                const SizedBox(height: 30),
                // Threat log
                Expanded(
                  child: Container(
                    width: double.infinity,
                    padding: const EdgeInsets.all(10),
                    decoration: BoxDecoration(border: Border.all(color: Colors.lime)),
                    child: ListView.builder(
                      itemCount: threats.length,
                      itemBuilder: (_, i) => Text(threats[i], style: const TextStyle(color: Colors.limeAccent, fontSize: 12)),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      ],
    ),
  );
}

// Simple matrix rain widget
class MatrixRain extends StatelessWidget {
  const MatrixRain({super.key});
  @override Widget build(BuildContext context) => CustomPaint(painter: _MatrixPainter());
}

class _MatrixPainter extends CustomPainter {
  @override void paint(Canvas canvas, Size size) {
    final paint = Paint()..color = Colors.lime.withOpacity(0.1);
    for (double x = 0; x < size.width; x += 20) {
      double y = DateTime.now().millisecond % 1000 / 1000 * size.height - 100;
      canvas.drawCircle(Offset(x, y), 2, paint);
    }
  }
  @override bool shouldRepaint(_) => true;
}